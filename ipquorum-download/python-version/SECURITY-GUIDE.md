# Security Guide: Password Management

This guide explains secure ways to handle passwords when using the IPQuorum download script.

## ⚠️ Security Risk Levels

| Method | Security Level | Visibility | Recommended For |
|--------|---------------|------------|-----------------|
| `--pass password` | ❌ **LOW** | Visible in command history, process list, logs | **Never use in production** |
| `--pass-prompt` | ✅ **HIGH** | Hidden input, not stored | **Interactive use (recommended)** |
| `--pass-file` | ✅ **HIGH** | File-based, can set permissions | **Automation (recommended)** |
| Environment variable | ⚠️ **MEDIUM** | Visible in process environment | **Temporary use only** |

---

## 🔒 Recommended Methods

### Method 1: Interactive Password Prompt (Most Secure for Manual Use)

The script will automatically prompt for password if none is provided:

```bash
python3 ipquorum-restapi-download.py \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --download --insecure

# Output:
# Enter password for superuser: ********
```

Or explicitly request prompt:

```bash
python3 ipquorum-restapi-download.py \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --pass-prompt \
  --download --insecure
```

**Advantages:**
- ✅ Password never appears in command history
- ✅ Password never appears in process list
- ✅ Hidden input (not visible while typing)
- ✅ Not stored anywhere

**Best for:** Interactive/manual script execution

---

### Method 2: Password File (Most Secure for Automation)

Store password in a file with restricted permissions:

```bash
# Create password file
echo 'your_secure_password' > ~/.ipquorum_password

# Set restrictive permissions (owner read-only)
chmod 400 ~/.ipquorum_password

# Use password file
python3 ipquorum-restapi-download.py \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --pass-file ~/.ipquorum_password \
  --download --insecure
```

**Advantages:**
- ✅ Password not in command line
- ✅ File permissions protect password
- ✅ Works in automation/scripts
- ✅ Can be managed by secrets management tools

**Best for:** Automated scripts, cron jobs, CI/CD pipelines

**Important:** Always set restrictive permissions:
```bash
chmod 400 password_file  # Owner read-only
# or
chmod 600 password_file  # Owner read-write
```

---

### Method 3: Environment Variable (Temporary Use)

```bash
# Set environment variable (current session only)
export VIRTUALIZE_PASSWORD='your_password'

# Run script (password not in command line)
python3 ipquorum-restapi-download.py \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --download --insecure

# Clear variable after use
unset VIRTUALIZE_PASSWORD
```

**Advantages:**
- ✅ Password not in command line
- ✅ Works in automation

**Disadvantages:**
- ⚠️ Visible in process environment (`/proc/<pid>/environ`)
- ⚠️ May be logged by shell history if set inline

**Best for:** Temporary use, testing

---

## ❌ Insecure Methods (Avoid in Production)

### Method 4: Command Line Password (INSECURE - DO NOT USE)

```bash
# ❌ INSECURE - Password visible everywhere
python3 ipquorum-restapi-download.py \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --pass 'your_password' \
  --download --insecure
```

**Why this is insecure:**
- ❌ Visible in command history (`~/.bash_history`)
- ❌ Visible in process list (`ps aux | grep ipquorum`)
- ❌ May be logged by system audit tools
- ❌ Visible to other users on the system

**Only acceptable for:** Testing in isolated environments

---

## 🛡️ Best Practices

### 1. Use Password Prompt for Interactive Use

```bash
# Best practice for manual execution
python3 ipquorum-restapi-download.py \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --pass-prompt \
  --download --insecure
```

### 2. Use Password File for Automation

```bash
# Create secure password file
echo 'password' > /secure/path/.ipquorum_pass
chmod 400 /secure/path/.ipquorum_pass

# Use in automation
python3 ipquorum-restapi-download.py \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --pass-file /secure/path/.ipquorum_pass \
  --download --insecure
```

### 3. Integrate with Secrets Management

#### HashiCorp Vault Example

```bash
# Retrieve password from Vault
PASSWORD=$(vault kv get -field=password secret/ipquorum)

# Use with environment variable (cleared after use)
VIRTUALIZE_PASSWORD="$PASSWORD" python3 ipquorum-restapi-download.py \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --download --insecure
```

#### AWS Secrets Manager Example

```bash
# Retrieve password from AWS Secrets Manager
PASSWORD=$(aws secretsmanager get-secret-value \
  --secret-id ipquorum-password \
  --query SecretString \
  --output text)

# Use with environment variable
VIRTUALIZE_PASSWORD="$PASSWORD" python3 ipquorum-restapi-download.py \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --download --insecure
```

### 4. Clear Shell History

If you accidentally used `--pass` in command line:

```bash
# Clear last command from history
history -d $(history 1 | awk '{print $1}')

# Or clear entire history
history -c
```

### 5. Use Restricted User Accounts

- Use **Monitor** role for download-only operations
- Use **Restricted Administrator** only when creating Quorum Apps
- Never use **superuser** unless absolutely necessary

---

## 🔍 Password Masking in Logs

The script automatically masks passwords in debug logs:

```bash
# Run with debug mode
python3 ipquorum-restapi-download.py --debug \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --pass-prompt \
  --download --insecure

# Debug output shows masked password:
# 2026-01-28 21:00:00 - DEBUG - user='superuser', endpoint='10.33.7.80'
# 2026-01-28 21:00:00 - DEBUG - password='********'
```

Passwords are **never** logged in plain text, even in debug mode.

---

## 📋 Security Checklist

Before running in production:

- [ ] **Never** use `--pass` with password in command line
- [ ] Use `--pass-prompt` for interactive use
- [ ] Use `--pass-file` with `chmod 400` for automation
- [ ] Store password files outside web-accessible directories
- [ ] Use secrets management tools (Vault, AWS Secrets Manager, etc.)
- [ ] Use least-privilege accounts (Monitor role when possible)
- [ ] Enable audit logging on the storage system
- [ ] Rotate passwords regularly
- [ ] Clear shell history if password was exposed
- [ ] Use `--secure` (TLS verification) in production

---

## 🚨 What to Do If Password Is Exposed

If you accidentally exposed a password:

1. **Change the password immediately** on the storage system
2. **Clear shell history:**
   ```bash
   history -c
   history -w
   ```
3. **Check for logged commands:**
   ```bash
   grep -r "your_password" ~/.bash_history ~/.zsh_history
   ```
4. **Review system logs** for unauthorized access
5. **Notify security team** if in production environment

---

## 📚 Additional Resources

- [IBM Storage Virtualize Security Best Practices](https://www.ibm.com/docs/en/flashsystem-9x00)
- [OWASP Password Storage Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Password_Storage_Cheat_Sheet.html)
- [CIS Benchmark for Password Management](https://www.cisecurity.org/)

---

## 🎯 Quick Reference

| Use Case | Recommended Method | Command |
|----------|-------------------|---------|
| Manual execution | `--pass-prompt` | `python3 script.py --pass-prompt ...` |
| Automation/cron | `--pass-file` | `python3 script.py --pass-file /path/to/file ...` |
| CI/CD pipeline | Environment variable + secrets manager | `VIRTUALIZE_PASSWORD=$(vault ...) python3 script.py ...` |
| Testing only | `--pass` (insecure) | `python3 script.py --pass test123 ...` |

---

**Remember:** Security is not optional. Always use secure password methods in production!