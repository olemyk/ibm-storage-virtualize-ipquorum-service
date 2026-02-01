# IPQuorum Download Scripts

Automated scripts to download the IBM Storage Virtualize IP Quorum JAR file via REST API and optionally create a new Quorum App.

---

## 📁 Available Versions

### 🐚 [Bash Version](bash-version/)
Traditional bash script with minimal dependencies.

**Best for:**
- Linux/Unix environments
- Minimal dependencies (bash, curl, jq)
- Simple automation tasks
- Users familiar with bash scripting

**Quick Start:**
```bash
cd bash-version
chmod +x ipquorum-restapi-download.sh
./ipquorum-restapi-download.sh --help
```

[📖 Bash Documentation](bash-version/README.md)

---

### 🐍 [Python Version](python-version/)
Modern Python implementation with enhanced features.

**Best for:**
- Cross-platform use (Windows, Linux, macOS)
- Secure password handling
- Advanced logging and debugging
- Enterprise environments
- Users who need password masking

**Quick Start:**
```bash
cd python-version
pip install -r requirements.txt
python3 ipquorum-restapi-download.py --help
```

[📖 Python Documentation](python-version/README.md)

---

## 🔄 Feature Comparison

| Feature | Bash | Python |
|---------|------|--------|
| **Platform** | Linux/macOS | Windows/Linux/macOS |
| **Dependencies** | bash, curl, jq | Python 3.6+, requests |
| **Fail-fast auth** | ✅ | ✅ |
| **Rate limiting** | ✅ | ✅ |
| **Password masking** | ✅ | ✅ |
| **Interactive prompt** | ✅ | ✅ |
| **Password files** | ✅ | ✅ |
| **Debug logging** | Basic | Advanced |
| **Type safety** | ❌ | ✅ |
| **Documentation** | Good | Comprehensive |

---

## 🚀 Quick Examples

### Bash Version
```bash
cd bash-version
./ipquorum-restapi-download.sh \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --pass password \
  --download --insecure
```

### Python Version (Secure)
```bash
cd python-version
python3 ipquorum-restapi-download.py \
  --api-endpoint 10.33.7.80 \
  --user superuser \
  --pass-prompt \
  --download --insecure
```

---

## 📚 Documentation

### Bash Version
- [README.md](bash-version/README.md) - Main documentation
- [simple-restapi-download.sh](bash-version/simple-restapi-download.sh) - Simple example

### Python Version
- [README.md](python-version/README.md) - Main documentation
- [INSTALL.md](python-version/INSTALL.md) - Installation guide
- [SECURITY-GUIDE.md](python-version/SECURITY-GUIDE.md) - Password security
- [WORKFLOW-DIAGRAM.md](python-version/WORKFLOW-DIAGRAM.md) - Visual workflows
- [MIGRATION-GUIDE.md](python-version/MIGRATION-GUIDE.md) - Bash to Python migration

### Manual Instructions
- [ipquorum-download-readme.md](ipquorum-download-readme.md) - Manual curl commands

---

## 🎯 Which Version Should I Use?

### Choose Bash if:
- ✅ You're on Linux/Unix only
- ✅ You want minimal dependencies
- ✅ You're comfortable with bash scripting
- ✅ You don't need password masking
- ✅ Simple automation is sufficient

### Choose Python if:
- ✅ You need cross-platform support (Windows)
- ✅ You want secure password handling
- ✅ You need password masking in logs
- ✅ You want interactive password prompts
- ✅ You prefer modern, maintainable code
- ✅ You need comprehensive documentation

---

## 🔒 Security Considerations

### Bash Version
- ✅ Interactive password prompt (hidden input)
- ✅ Password file support with permissions
- ✅ Password masking in debug output
- ✅ Multiple secure input methods
- ⚠️ Basic password masking (shows *******)

### Python Version
- ✅ Interactive password prompt (hidden input)
- ✅ Password file support with permissions
- ✅ Password masking in all logs
- ✅ Multiple secure input methods
- ✅ Advanced password masking with options

See [Python Security Guide](python-version/SECURITY-GUIDE.md) for best practices.

---

## 📖 IBM Documentation

- [IPQuorum Info](https://www.ibm.com/support/pages/ibm-storage-virtualize-ip-quorum-application-requirements-1)
- [IP quorum application](https://www.ibm.com/docs/en/flashsystem-9x00/9.1.1?topic=quorum-ip-application)
- [Storage Virtualize RESTful API](https://www.ibm.com/docs/en/flashsystem-9x00/9.1.1?topic=interface-storage-virtualize-restful-api)

---

## 🤝 Contributing

Both versions are maintained and improvements are welcome. Please ensure:
- Bash version maintains POSIX compatibility
- Python version maintains Python 3.6+ compatibility
- Documentation is updated for any changes

---

## 👤 Maintainer
Ole Kristian Myklebust

---

## 📄 License
MIT License