#!/usr/bin/env python3
"""
IBM Storage Virtualize IP Quorum Download Script

A Python utility to download the IBM Virtualize IPQuorum JAR via REST API
and optionally create a new Quorum App (mkquorumapp).

Author: Ole Kristian Myklebust
Python Version: 3.6+ (3.8+ recommended)
"""

import argparse
import getpass
import json
import logging
import os
import random
import sys
import time
from dataclasses import dataclass
from typing import Optional, Dict, Any, Tuple

import requests
from requests.adapters import HTTPAdapter
from urllib3.util.retry import Retry
import urllib3

# Disable SSL warnings when using insecure mode
urllib3.disable_warnings(urllib3.exceptions.InsecureRequestWarning)


# ===== Custom Exceptions =====
class IPQuorumError(Exception):
    """Base exception for IPQuorum operations"""
    pass


class AuthenticationError(IPQuorumError):
    """Authentication failed"""
    pass


class APIError(IPQuorumError):
    """API call failed"""
    pass


class ValidationError(IPQuorumError):
    """Configuration validation failed"""
    pass


class NetworkError(IPQuorumError):
    """Network connectivity issue"""
    pass


# ===== Configuration =====
@dataclass
class Config:
    """Configuration for IPQuorum operations"""
    api_endpoint: str
    username: str
    password: str
    api_port: int = 7443
    output_file: str = "ip_quorum.jar"
    verify_ssl: bool = False
    mkquorumapp: bool = True
    download: bool = True
    max_retries: int = 8
    base_delay: int = 3
    
    # mkquorumapp parameters
    ip6: bool = False
    nometadata: bool = False
    partnersystem: str = ""
    partnerip6: bool = False
    
    def validate(self) -> None:
        """Validate configuration parameters"""
        if not self.password:
            raise ValidationError("--pass <password> is required")
        
        if self.mkquorumapp and not self.partnersystem:
            raise ValidationError(
                "--partnersystem <name> is MANDATORY when --mkquorumapp is enabled"
            )
        
        if not self.api_endpoint:
            raise ValidationError("--api-endpoint <host> is required")


# ===== Helper Functions =====
def mask_password(password: str, show_chars: int = 0) -> str:
    """
    Mask password for logging purposes.
    
    Args:
        password: Password to mask
        show_chars: Number of characters to show at the end (default: 0)
        
    Returns:
        str: Masked password
    """
    if not password:
        return "<empty>"
    
    if len(password) <= show_chars:
        return "*" * len(password)
    
    if show_chars > 0:
        return "*" * (len(password) - show_chars) + password[-show_chars:]
    else:
        return "*" * min(len(password), 8)


def read_password_from_file(filepath: str) -> str:
    """
    Read password from a file.
    
    Args:
        filepath: Path to password file
        
    Returns:
        str: Password from file
        
    Raises:
        ValidationError: If file cannot be read
    """
    try:
        with open(filepath, 'r') as f:
            password = f.read().strip()
        return password
    except FileNotFoundError:
        raise ValidationError(f"Password file not found: {filepath}")
    except PermissionError:
        raise ValidationError(f"Permission denied reading password file: {filepath}")
    except Exception as e:
        raise ValidationError(f"Error reading password file: {e}")


def get_password_interactive(username: str) -> str:
    """
    Prompt user for password interactively (hidden input).
    
    Args:
        username: Username for the prompt
        
    Returns:
        str: Password entered by user
    """
    try:
        password = getpass.getpass(f"Enter password for {username}: ")
        return password
    except KeyboardInterrupt:
        print("\nPassword input cancelled")
        sys.exit(130)


def to_bool(value: Any) -> bool:
    """
    Convert various inputs to boolean.
    
    Args:
        value: Input value to convert
        
    Returns:
        bool: Converted boolean value
    """
    if isinstance(value, bool):
        return value
    
    if value is None:
        return False
    
    str_value = str(value).lower().strip()
    
    if str_value in ('true', '1', 'yes', 'y', 'on'):
        return True
    elif str_value in ('false', '0', 'no', 'n', 'off', ''):
        return False
    else:
        return False


def setup_logging(debug: bool = False) -> logging.Logger:
    """
    Configure logging system.
    
    Args:
        debug: Enable debug level logging
        
    Returns:
        logging.Logger: Configured logger
    """
    level = logging.DEBUG if debug else logging.INFO
    
    logging.basicConfig(
        level=level,
        format='%(asctime)s - %(levelname)s - %(message)s',
        datefmt='%Y-%m-%d %H:%M:%S'
    )
    
    logger = logging.getLogger('ipquorum')
    return logger


# ===== IPQuorum Client =====
class IPQuorumClient:
    """Client for IBM Storage Virtualize REST API operations"""
    
    def __init__(self, config: Config, logger: logging.Logger):
        """
        Initialize IPQuorum client.
        
        Args:
            config: Configuration object
            logger: Logger instance
        """
        self.config = config
        self.logger = logger
        self.access_token: Optional[str] = None
        self.base_url = f"https://{config.api_endpoint}:{config.api_port}"
        
        # Create session with retry strategy
        self.session = requests.Session()
        
        # Configure session
        self.session.verify = config.verify_ssl
        
        self.logger.debug(f"Initialized client for {self.base_url}")
        self.logger.debug(f"SSL verification: {config.verify_ssl}")
    
    def preflight_check(self) -> bool:
        """
        Validate endpoint reachability.
        
        Returns:
            bool: True if endpoint is reachable
            
        Raises:
            NetworkError: If endpoint is unreachable
        """
        self.logger.info("Pre-flight: validating inputs & endpoint reachability...")
        
        url = f"{self.base_url}/rest/v1/auth"
        
        try:
            response = self.session.get(
                url,
                timeout=10,
                verify=self.config.verify_ssl
            )
            
            http_code = response.status_code
            self.logger.info(f"Pre-flight: endpoint status: {http_code}")
            
            # Log response headers
            self.logger.debug("Pre-flight: endpoint headers:")
            for key, value in response.headers.items():
                self.logger.debug(f"  {key}: {value}")
            
            # Accept various status codes that indicate the endpoint is reachable
            if http_code in (200, 401, 404, 405):
                self.logger.info("Pre-flight: endpoint is reachable.")
                return True
            else:
                self.logger.warning(
                    f"Pre-flight: unexpected status {http_code}. Proceeding may fail."
                )
                return True
                
        except requests.exceptions.SSLError as e:
            self.logger.error(
                f"Pre-flight: SSL/TLS error. Try --insecure or verify connectivity."
            )
            raise NetworkError(f"SSL/TLS error: {e}")
        except requests.exceptions.ConnectionError as e:
            self.logger.error(
                f"Pre-flight: Connection error. Verify endpoint and network connectivity."
            )
            raise NetworkError(f"Connection error: {e}")
        except requests.exceptions.Timeout as e:
            self.logger.error("Pre-flight: Request timeout.")
            raise NetworkError(f"Timeout error: {e}")
        except Exception as e:
            self.logger.error(f"Pre-flight: Unexpected error: {e}")
            raise NetworkError(f"Unexpected error: {e}")
    
    def authenticate(self) -> str:
        """
        Authenticate with retry logic and rate limiting.
        
        Returns:
            str: Authentication token
            
        Raises:
            AuthenticationError: If authentication fails after all retries
        """
        self.logger.info("Get Token, please wait...")
        
        url = f"{self.base_url}/rest/v1/auth"
        headers = {
            "Accept": "application/json",
            "Content-Type": "application/json",
            "X-Auth-Username": self.config.username,
            "X-Auth-Password": self.config.password
        }
        
        for attempt in range(1, self.config.max_retries + 1):
            self.logger.info(f"Auth attempt {attempt}/{self.config.max_retries}...")
            
            try:
                response = self.session.post(
                    url,
                    headers=headers,
                    timeout=30,
                    verify=self.config.verify_ssl
                )
                
                http_code = response.status_code
                self.logger.debug(f"Auth response status: {http_code}")
                
                # Don't retry on authentication failures (wrong credentials)
                if http_code == 401:
                    self.logger.error("Authentication failed: Invalid username or password")
                    raise AuthenticationError("Invalid username or password (HTTP 401)")
                
                # Don't retry on forbidden (insufficient permissions)
                if http_code == 403:
                    self.logger.error("Authentication failed: Insufficient permissions or invalid credentials")
                    raise AuthenticationError("Insufficient permissions or invalid credentials (HTTP 403)")
                
                # Check for rate limiting
                if http_code == 429:
                    retry_after = response.headers.get('Retry-After')
                    if retry_after:
                        delay = int(retry_after)
                    else:
                        delay = self.config.base_delay * attempt + random.randint(0, 2)
                    
                    self.logger.warning(
                        f"Rate limited (429). Sleeping {delay}s and retrying..."
                    )
                    time.sleep(delay)
                    continue
                
                # Try to extract token from headers
                token = (
                    response.headers.get('X-Auth-Token') or
                    response.headers.get('x-auth-token') or
                    response.headers.get('Authorization') or
                    response.headers.get('authorization')
                )
                
                # If not in headers, try JSON body
                if not token or token == 'null':
                    try:
                        body = response.json()
                        if isinstance(body, dict):
                            token = (
                                body.get('token') or
                                body.get('access_token') or
                                body.get('authToken') or
                                body.get('session') or
                                (body.get('data', {}).get('token') if isinstance(body.get('data'), dict) else None) or
                                (body.get('result', {}).get('token') if isinstance(body.get('result'), dict) else None)
                            )
                    except (json.JSONDecodeError, AttributeError):
                        pass
                
                # Check if authentication was successful
                if http_code in (200, 201) and token and token != 'null':
                    self.logger.info("Authentication successful")
                    self.access_token = token
                    return token
                
                # Log failure details for retryable errors
                self.logger.warning(
                    f"Attempt {attempt} failed (http_code={http_code}, "
                    f"token={'found' if token else 'not found'}). Retrying..."
                )
                
                # Wait before retry
                delay = self.config.base_delay + attempt + random.randint(0, 1)
                time.sleep(delay)
                
            except requests.exceptions.RequestException as e:
                self.logger.warning(f"Attempt {attempt} failed with exception: {e}")
                delay = self.config.base_delay + attempt + random.randint(0, 1)
                time.sleep(delay)
        
        raise AuthenticationError(
            f"Failed to obtain token after {self.config.max_retries} attempts"
        )
    
    def create_quorum_app(
        self,
        ip6: bool,
        nometadata: bool,
        partnersystem: str,
        partnerip6: bool
    ) -> Dict[str, Any]:
        """
        Create IP Quorum application via mkquorumapp API.
        
        Args:
            ip6: IPv6 flag for local system
            nometadata: Metadata flag
            partnersystem: Partner system name (remote system in PBHA)
            partnerip6: IPv6 flag for partner system
            
        Returns:
            dict: API response
            
        Raises:
            APIError: If API call fails
        """
        if not self.access_token:
            raise APIError("Not authenticated. Call authenticate() first.")
        
        self.logger.info("Creating IP-Quorum app via /rest/v1/mkquorumapp...")
        
        url = f"{self.base_url}/rest/v1/mkquorumapp"
        
        payload = {
            "ip_6": ip6,
            "nometadata": nometadata,
            "partnersystem": partnersystem,
            "partnerip6": partnerip6
        }
        
        self.logger.debug(f"Payload: {json.dumps(payload)}")
        
        headers = {
            "accept": "application/json",
            "X-Auth-Token": self.access_token,
            "Content-Type": "application/json"
        }
        
        try:
            response = self.session.post(
                url,
                headers=headers,
                json=payload,
                timeout=60,
                verify=self.config.verify_ssl
            )
            
            http_code = response.status_code
            self.logger.info(f"mkquorumapp response status: {http_code}")
            
            # Log response
            try:
                response_data = response.json()
                self.logger.debug(f"Response: {json.dumps(response_data, indent=2)}")
            except json.JSONDecodeError:
                self.logger.debug(f"Response text: {response.text}")
            
            if http_code in (200, 201):
                self.logger.info("mkquorumapp call completed successfully")
                try:
                    return response.json()
                except json.JSONDecodeError:
                    return {"status": "success", "text": response.text}
            else:
                error_msg = f"mkquorumapp failed with status {http_code}"
                self.logger.error(error_msg)
                raise APIError(error_msg)
                
        except requests.exceptions.RequestException as e:
            error_msg = f"mkquorumapp request failed: {e}"
            self.logger.error(error_msg)
            raise APIError(error_msg)
    
    def download_jar(self, output_file: str) -> None:
        """
        Download ip_quorum.jar file.
        
        Args:
            output_file: Output filename
            
        Raises:
            APIError: If download fails
        """
        if not self.access_token:
            raise APIError("Not authenticated. Call authenticate() first.")
        
        self.logger.info(f"Proceeding to download {output_file}...")
        
        url = f"{self.base_url}/rest/v1/download"
        
        payload = {
            "prefix": "/dumps",
            "filename": "ip_quorum.jar"
        }
        
        headers = {
            "accept": "application/json",
            "X-Auth-Token": self.access_token,
            "Content-Type": "application/json"
        }
        
        try:
            response = self.session.post(
                url,
                headers=headers,
                json=payload,
                timeout=300,
                stream=True,
                verify=self.config.verify_ssl
            )
            
            http_code = response.status_code
            self.logger.info(f"Download response status: {http_code}")
            
            # Log response headers
            self.logger.debug("Download Headers:")
            for key, value in response.headers.items():
                self.logger.debug(f"  {key}: {value}")
            
            if http_code in (200, 201):
                # Write file
                with open(output_file, 'wb') as f:
                    for chunk in response.iter_content(chunk_size=8192):
                        if chunk:
                            f.write(chunk)
                
                # Verify file was created and has content
                if os.path.exists(output_file):
                    file_size = os.path.getsize(output_file)
                    if file_size > 0:
                        self.logger.info(
                            f"Downloaded {output_file} (size: {file_size} bytes)"
                        )
                    else:
                        self.logger.warning(
                            "Success status but empty file. Check server response."
                        )
                else:
                    raise APIError("File was not created")
            else:
                error_msg = f"Download failed with status {http_code}"
                self.logger.error(error_msg)
                raise APIError(error_msg)
                
        except requests.exceptions.RequestException as e:
            error_msg = f"Download request failed: {e}"
            self.logger.error(error_msg)
            raise APIError(error_msg)


# ===== Argument Parsing =====
def parse_arguments() -> argparse.Namespace:
    """
    Parse command-line arguments.
    
    Returns:
        argparse.Namespace: Parsed arguments
    """
    parser = argparse.ArgumentParser(
        description='IBM Storage Virtualize IP Quorum Download Script',
        formatter_class=argparse.RawDescriptionHelpFormatter,
        epilog='''
Examples:
  # Interactive password prompt (most secure for manual use)
  %(prog)s --api-endpoint 10.33.7.80 \\
    --user superuser --pass-prompt \\
    --download --insecure

  # Password from file (most secure for automation)
  %(prog)s --api-endpoint 10.33.7.80 \\
    --user superuser --pass-file ~/.ipquorum_password \\
    --download --insecure

  # Create Quorum App + Download JAR
  %(prog)s --api-endpoint 10.33.7.80 \\
    --mkquorumapp --partnersystem svc_cluster02 \\
    --ip6=false --partnerip6=false --nometadata=false \\
    --download --insecure --user superuser --pass-prompt

  # Use environment variables
  export API_ENDPOINT=10.33.7.80
  export VIRTUALIZE_USERNAME=superuser
  %(prog)s --download --insecure --pass-prompt

Security:
  See SECURITY-GUIDE.md for password best practices.
  Never use --pass with password in command line in production!
        '''
    )
    
    # General options
    parser.add_argument(
        '--mkquorumapp',
        dest='mkquorumapp',
        action='store_true',
        default=None,
        help='Enable mkquorumapp call (default: enabled)'
    )
    parser.add_argument(
        '--no-mkquorumapp',
        dest='mkquorumapp',
        action='store_false',
        help='Disable mkquorumapp call'
    )
    
    parser.add_argument(
        '--download',
        dest='download',
        action='store_true',
        default=None,
        help='Enable jar download (default: enabled)'
    )
    parser.add_argument(
        '--no-download',
        dest='download',
        action='store_false',
        help='Disable jar download'
    )
    
    parser.add_argument(
        '--insecure',
        dest='insecure',
        action='store_true',
        default=True,
        help='Use insecure TLS (default)'
    )
    parser.add_argument(
        '--secure',
        dest='insecure',
        action='store_false',
        help='Use strict TLS verification'
    )
    
    parser.add_argument(
        '--api-endpoint',
        type=str,
        default=os.environ.get('API_ENDPOINT', ''),
        help='API endpoint hostname/IP (env: API_ENDPOINT)'
    )
    
    parser.add_argument(
        '--user',
        type=str,
        default=os.environ.get('VIRTUALIZE_USERNAME', ''),
        help='Auth username (env: VIRTUALIZE_USERNAME)'
    )
    
    # Password options (mutually exclusive group)
    password_group = parser.add_mutually_exclusive_group()
    password_group.add_argument(
        '--pass',
        dest='password',
        type=str,
        default=os.environ.get('VIRTUALIZE_PASSWORD', ''),
        help='Auth password (env: VIRTUALIZE_PASSWORD)'
    )
    password_group.add_argument(
        '--pass-file',
        dest='password_file',
        type=str,
        help='Read password from file (more secure than --pass)'
    )
    password_group.add_argument(
        '--pass-prompt',
        dest='password_prompt',
        action='store_true',
        help='Prompt for password interactively (hidden input, most secure)'
    )
    
    parser.add_argument(
        '--output',
        type=str,
        default=os.environ.get('IPQ_OUTPUT_FILE', 'ip_quorum.jar'),
        help='Output jar filename (default: ip_quorum.jar)'
    )
    
    # mkquorumapp payload options
    parser.add_argument(
        '--ip6', '--ip_6',
        type=lambda x: to_bool(x) if x else None,
        nargs='?',
        const=True,
        default=None,
        help='Set IPv6 flag (default: false)'
    )
    
    parser.add_argument(
        '--nometadata',
        type=lambda x: to_bool(x) if x else None,
        nargs='?',
        const=True,
        default=None,
        help='Set nometadata flag (default: false)'
    )
    
    parser.add_argument(
        '--partnersystem',
        type=str,
        default=os.environ.get('partnersystem', ''),
        help='Set Partnersystem - Remote System in PBHA (MANDATORY if mkquorumapp enabled)'
    )
    
    parser.add_argument(
        '--partnerip6',
        type=lambda x: to_bool(x) if x else None,
        nargs='?',
        const=True,
        default=None,
        help='Set partner IPv6 flag (default: false)'
    )
    
    parser.add_argument(
        '--debug',
        action='store_true',
        help='Enable debug logging'
    )
    
    return parser.parse_args()


# ===== Main Function =====
def main() -> int:
    """
    Main execution function.
    
    Returns:
        int: Exit code (0=success, 1=error, 2=validation error, 3=network error)
    """
    # Initialize logger early to ensure it's available in exception handlers
    logger = setup_logging(debug=False)
    
    try:
        # Parse arguments
        args = parse_arguments()
        
        # Update logging level if debug is enabled
        if args.debug:
            logger.setLevel(logging.DEBUG)
            for handler in logger.handlers:
                handler.setLevel(logging.DEBUG)
        
        # Handle password input methods
        password = args.password
        
        if args.password_file:
            logger.info(f"Reading password from file: {args.password_file}")
            password = read_password_from_file(args.password_file)
        elif args.password_prompt:
            password = get_password_interactive(args.user)
        elif not password:
            # No password provided via any method, prompt interactively
            logger.info("No password provided, prompting interactively...")
            password = get_password_interactive(args.user)
        
        # Handle defaults for boolean flags
        mkquorumapp = args.mkquorumapp if args.mkquorumapp is not None else True
        download = args.download if args.download is not None else True
        ip6 = args.ip6 if args.ip6 is not None else False
        nometadata = args.nometadata if args.nometadata is not None else False
        partnerip6 = args.partnerip6 if args.partnerip6 is not None else False
        
        # Debug output (with masked password)
        logger.debug(f"mkquorumapp={mkquorumapp}, download={download}, insecure={args.insecure}")
        logger.debug(f"ip6={ip6}, nometadata={nometadata}, partnersystem='{args.partnersystem}', partnerip6={partnerip6}")
        logger.debug(f"user='{args.user}', endpoint='{args.api_endpoint}'")
        logger.debug(f"password='{mask_password(password)}'")
        
        # Create configuration
        config = Config(
            api_endpoint=args.api_endpoint,
            username=args.user,
            password=password,
            output_file=args.output,
            verify_ssl=not args.insecure,
            mkquorumapp=mkquorumapp,
            download=download,
            ip6=ip6,
            nometadata=nometadata,
            partnersystem=args.partnersystem,
            partnerip6=partnerip6
        )
        
        # Validate configuration
        config.validate()
        
        # Initialize client
        client = IPQuorumClient(config, logger)
        
        # Pre-flight check
        client.preflight_check()
        
        # Authenticate
        client.authenticate()
        
        # Create quorum app (if enabled)
        if config.mkquorumapp:
            client.create_quorum_app(
                ip6=config.ip6,
                nometadata=config.nometadata,
                partnersystem=config.partnersystem,
                partnerip6=config.partnerip6
            )
        else:
            logger.info("Condition not met. Skipping the Create new IP-Quorum app call.")
        
        # Download JAR (if enabled)
        if config.download:
            client.download_jar(config.output_file)
        else:
            logger.info("Condition not met. Skipping the downloading of quorumapp.")
        
        logger.info("Operation completed successfully")
        return 0
        
    except ValidationError as e:
        logger.error(f"Validation error: {e}")
        return 2
    except NetworkError as e:
        logger.error(f"Network error: {e}")
        return 3
    except (AuthenticationError, APIError) as e:
        logger.error(f"API error: {e}")
        return 1
    except KeyboardInterrupt:
        logger.info("Operation cancelled by user")
        return 130
    except Exception as e:
        logger.error(f"Unexpected error: {e}", exc_info=True)
        return 1


if __name__ == "__main__":
    sys.exit(main())

# Made with help from Bob
