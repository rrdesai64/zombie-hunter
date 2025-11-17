# 🧟 Zombie Hunter

# 🧟 Zombie Hunter

[![Tests](https://github.com/rrdesai64/zombie-hunter/actions/workflows/test.yml/badge.svg)](https://github.com/rrdesai64/zombie-hunter/actions/workflows/test.yml)

**Find and eliminate zombie CronJobs in your Kubernetes clusters.**
**Find and eliminate zombie CronJobs in your Kubernetes clusters.**

Zombie CronJobs are scheduled jobs that haven't run successfully in weeks or months, yet they're still deployed and costing you money. Zombie Hunter finds them automatically.

---

## 🎯 What It Does

- ✅ Scans your Kubernetes cluster for CronJobs
- ✅ Identifies jobs that haven't run successfully recently
- ✅ Calculates confidence scores
- ✅ Exports reports in multiple formats (table, CSV, JSON)
- ✅ Helps you clean up and save money

---

## 🚀 Quick Start

### Installation

**Option 1: Download Binary** (Coming Soon)
```bash
# Releases will be available soon
```

**Option 2: Build from Source**
```bash
# Clone the repository
git clone https://github.com/rrdesai64/zombie-hunter
cd zombie-hunter

# Build
go build -o zombie-hunter.exe ./cmd/zombie-hunter

# Run
.\zombie-hunter.exe --help
```

---

## 📖 Usage
```bash
# Find zombies (default: 30 days inactive)
.\zombie-hunter.exe

# Conservative approach (60 days)
.\zombie-hunter.exe --days 60

# Aggressive cleanup (7 days)
.\zombie-hunter.exe --days 7

# Specific namespace only
.\zombie-hunter.exe --namespace production

# Export to CSV
.\zombie-hunter.exe --format csv > zombies.csv

# Export to JSON
.\zombie-hunter.exe --format json > zombies.json
```

---

## 📊 Example Output
```
🧟 ZOMBIE HUNTER REPORT
Generated: 2024-11-13 20:30:00
Threshold: 30 days

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
ZOMBIE CANDIDATES (3 found)
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

🔍   NAME                           NAMESPACE       DAYS INACTIVE   CONFIDENCE    JOBS
------------------------------------------------------------------------------------------------
💀   old-backup-job                 default         127             95%           5 total, 0 failed
⚠️   deprecated-cleanup             staging         45              75%           12 total, 3 failed
🤔   experimental-task              dev             15              50%           2 total, 0 failed

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
SUMMARY
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━

Total zombies found: 3
High confidence (≥80%): 1

💡 Tip: Start by reviewing high-confidence zombies

Next steps:
1. Review each zombie with your team
2. Delete safely: kubectl delete cronjob <name> -n <namespace>
3. Try different thresholds: --days 60 or --days 90
```

---

## 🎯 Why Zombie Hunter?

**The Problem:**
- DevOps teams create hundreds of CronJobs
- Many become abandoned over time
- They cost money (compute resources)
- They create security risks (unmaintained code)
- Nobody dares to delete them (fear of breaking production)

**The Solution:**
- Zombie Hunter identifies them automatically
- Provides confidence scores
- Makes cleanup safe and easy

---

## 🗺️ Roadmap

- [x] Basic zombie detection
- [x] Multiple output formats (table, CSV, JSON)
- [x] Configurable thresholds
- [x] Namespace filtering
- [ ] Safe delete with automatic rollback
- [ ] Web dashboard
- [ ] Multi-cluster support
- [ ] Slack/Email notifications
- [ ] Cost estimation
- [ ] Historical tracking

---

## 🤝 Contributing

Contributions are welcome! This project is in early development.

To contribute:
1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Submit a pull request

---

## 📝 License

MIT License - see [LICENSE](LICENSE) file for details.

---

## 👤 Author

**Rama Rao Desai**
- GitHub: [@rrdesai64](https://github.com/rrdesai64)

---

## ⚠️ Status

**🚧 Work in Progress**

This project is actively being developed. Features and APIs may change.

Current version: **v0.1.0-alpha**

---

## 💬 Feedback

Found a bug? Have a feature request? 

[Open an issue](https://github.com/rrdesai64/zombie-hunter/issues) on GitHub!

---

**Made with ❤️ for DevOps teams everywhere**