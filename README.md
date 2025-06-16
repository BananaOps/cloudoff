# Cloudoff

**Cloudoff** is a lightweight Go application designed to automate the management of AWS EC2 instances. It helps reduce cloud costs by automatically stopping instances at night, restarting them during the day, and cleaning up instances that have exceeded their Time-To-Live (TTL).

## 🚀 Features

- ⏰ **Scheduled stop and start** of EC2 instances (e.g., stop at night, start in the morning)
- 🧹 **Automatic cleanup** of EC2 instances after their **TTL** expiration
- 📊 Optimizes cost by limiting the runtime of non-critical resources
- 🔒 Follows AWS best practices (tag-based selection, IAM roles, etc.)

## 🔧 Requirements

- Go 1.18+
- AWS credentials with EC2 management permissions (via environment or IAM role)
- Tagged EC2 instances (see below)

## 🛠️ Installation

```bash
git clone https://github.com/bananaops/cloudoff.git
cd cloudoff
go build -o cloudoff

```

## ⚙️ Usage

### 🏷️ EC2 Tags Used by Cloudoff

Cloudoff relies on specific EC2 tags to determine which instances to manage and when to clean them up.

| Tag Key              | Example Value              | Description                                                                 |
|----------------------|----------------------------|-----------------------------------------------------------------------------|
| `cloudoff:uptime`    | `Mon-Fri 08:00-20:00 Europe/Paris`         | Specifies when the instance should be running. Timezone must be specified.      |
| `cloudoff:downtime`  | `Sat-Sun 00:00-23:59 Europe/Paris`         | Specifies when the instance must be stopped. Overrides `uptime` if both overlap.|
| `cloudoff:ttl`       | `3d` or `12h` or `1w`                      | Time-to-live from instance launch. Supports `h` (hours), `d` (days), `w` (weeks).|

*ttl starts counting from instance atttach time of first network insterface. If exceeded, the instance is considered expired and eligible for termination.
