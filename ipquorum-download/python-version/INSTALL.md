# Installation Guide

## Quick Install

For Python 3.7+ (recommended):
```bash
pip install -r requirements.txt
```

For Python 3.6:
```bash
pip install -r requirements.txt
# This will automatically install the dataclasses backport
```

## Detailed Installation Steps

### 1. Check Python Version

```bash
python3 --version
```

**Minimum required**: Python 3.6  
**Recommended**: Python 3.8 or higher

### 2. Install Dependencies

#### Option A: Using requirements.txt (Recommended)
```bash
cd ipquorum-download
pip install -r requirements.txt
```

#### Option B: Manual Installation
```bash
# For Python 3.7+
pip install requests>=2.27.0 urllib3>=1.26.0

# For Python 3.6 (requires dataclasses backport)
pip install requests>=2.27.0 urllib3>=1.26.0 dataclasses>=0.6
```

### 3. Make Script Executable (Linux/macOS)

```bash
chmod +x ipquorum-restapi-download.py
```

### 4. Verify Installation

```bash
python3 ipquorum-restapi-download.py --help
```

You should see the help message without any errors.

## Python 3.6 Specific Notes

If you're using Python 3.6, the `dataclasses` module is not included in the standard library. The `requirements.txt` file automatically handles this by installing the `dataclasses` backport package.

### Troubleshooting Python 3.6

If you encounter `ModuleNotFoundError: No module named 'dataclasses'`:

```bash
# Install the dataclasses backport explicitly
pip install dataclasses

# Or upgrade pip and reinstall
pip install --upgrade pip
pip install -r requirements.txt
```

## Platform-Specific Instructions

### Linux (RHEL/CentOS 7 with Python 3.6)

```bash
# Install Python 3.6 if not available
sudo yum install python3 python3-pip

# Install dependencies
pip3 install -r requirements.txt

# Make executable
chmod +x ipquorum-restapi-download.py
```

### Linux (Ubuntu/Debian)

```bash
# Install Python 3 and pip
sudo apt-get update
sudo apt-get install python3 python3-pip

# Install dependencies
pip3 install -r requirements.txt

# Make executable
chmod +x ipquorum-restapi-download.py
```

### macOS

```bash
# Install Python 3 via Homebrew (if needed)
brew install python3

# Install dependencies
pip3 install -r requirements.txt

# Make executable
chmod +x ipquorum-restapi-download.py
```

### Windows

```powershell
# Install dependencies
pip install -r requirements.txt

# Run script
python ipquorum-restapi-download.py --help
```

## Virtual Environment (Recommended)

Using a virtual environment is recommended to avoid conflicts with system packages:

```bash
# Create virtual environment
python3 -m venv venv

# Activate virtual environment
# Linux/macOS:
source venv/bin/activate
# Windows:
venv\Scripts\activate

# Install dependencies
pip install -r requirements.txt

# Run script
python ipquorum-restapi-download.py --help

# Deactivate when done
deactivate
```

## Upgrading

To upgrade to the latest version of dependencies:

```bash
pip install --upgrade -r requirements.txt
```

## Uninstall

To remove the installed packages:

```bash
pip uninstall -y requests urllib3 dataclasses
```

## Common Issues

### Issue: pip not found
**Solution**: Install pip
```bash
# Linux
sudo apt-get install python3-pip  # Debian/Ubuntu
sudo yum install python3-pip      # RHEL/CentOS

# macOS
python3 -m ensurepip --upgrade
```

### Issue: Permission denied when installing
**Solution**: Use user installation
```bash
pip install --user -r requirements.txt
```

### Issue: Old pip version
**Solution**: Upgrade pip
```bash
pip install --upgrade pip
```

### Issue: SSL certificate errors
**Solution**: Upgrade certifi
```bash
pip install --upgrade certifi
```

## Verification

After installation, verify everything works:

```bash
# Test import
python3 -c "import requests; import dataclasses; print('All dependencies OK')"

# Test script
python3 ipquorum-restapi-download.py --help
```

If both commands succeed, you're ready to use the script!

## Next Steps

- Read [README-python.md](README-python.md) for usage instructions
- Check [MIGRATION-GUIDE.md](MIGRATION-GUIDE.md) if migrating from bash version
- Review examples in the documentation

## Support

For issues or questions:
- Check the troubleshooting section in README-python.md
- Verify Python version compatibility
- Ensure all dependencies are installed correctly