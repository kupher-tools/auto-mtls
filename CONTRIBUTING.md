# Contributing to Auto-mTLS Operator 🚀

First off — thanks for considering contributing! 🎉 Contributions are what make the open-source community amazing, so we’d love your help.

There are many ways you can contribute: reporting bugs, suggesting features, improving docs, or sending pull requests.

## 🛠️ Getting Started
### 1. Fork & Clone
- Fork a auto-mtls repo : https://github.com/kupher-tools/auto-mtls.git
  
- Clone repo to your machine or Github Codespace : `git clone https://github.com/<your-username>/auto-mtls.git`


### 2. Set Up Dev Environment

You’ll need:

 ```sh
Go v1.24.0+

Docker 17.03+

Kubectl v1.11.3+

Kubernetes v1.30+

Operator SDK v1.41.1
```

### 3. Build & Deploy Locally
- Run locally using `make run` command
- Build docker image using `make docker-build docker-push IMG=<your-registry>/auto-mtls:dev`
- Deploy complete operator on your K8s cluster `make deploy IMG=<your-registry>/auto-mtls:dev`

## 📌 Ways to Contribute
### 🐛 Report Bugs

- Use GitHub Issues.

- Include steps to reproduce, logs, and your cluster setup.

### 💡 Suggest Features

- Open an issue labeled enhancement.

- Describe the problem and your proposed solution.

### 📝 Improve Documentation

- Fix typos, clarify instructions, or add examples.

- PRs for README.md, CONTRIBUTING.md, or docs/ are always welcome.

### 💻 Submit Code

- Create a new branch from main.

- Follow Go best practices & run make test before pushing.

- Open a Pull Request with a clear description of your changes.

### 🔍 Pull Request Guidelines

- Keep PRs focused — small, logical chunks are easier to review.

- Include tests where applicable.

- Update docs (README.md, examples/) if your change affects users.

- Ensure make test passes.

### ⭐ Recognition

We use the All Contributors spec to recognize all forms of contributions.
Everyone who helps — code, docs, issues, reviews — will show up in our Contributors section 🙌

### 📚 Resources

- Operator Docs

- cert-manager


📄 License

By contributing, you agree that your contributions will be licensed under the Apache 2.0 License.
