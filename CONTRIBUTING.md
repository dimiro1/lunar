# Contributing

Thanks for want to contribute to FaaS-Go! We value any type of contribution, whether it's reporting bugs, suggesting features, or submitting code changes.

## Roadmap

There are a few things that I want to implement in the near future:

### Quick wins

- Make the admin application responsive for mobile devices.
- Add the chance to maximize the code editor on the admin application.
  - We can create a button that toggles the maximization of the code editor.
- Cleanup the codebase by separating the API types from the database types.
- Setting up goreleaser.
- Vendor JS dependencies
  - Monaco editor
  - Highlight.js
  - Mithril.js
- One click install
  - Railway
  - Fly
  - coolify
- Accessiblity ARIA improvements on the admin application.

### Nice additions

- Implement a dashbard with real-time metrics and function monitoring.
- Add end to end tests for the entire system. Specially to test the admin application.
- Implement a blob storage API.
  - For this one, I want to introduce a bucket list interface where the user can list the files, view them, delete them, etc.
  - We offer a a new set of lua apis to interact with the blob storage.
- In memory caching with automatic invalidation.
- Cron scheduling for functions.
- Job queues and background processing.
- AI integration enhancements
  - Add support for more AI providers (Ollama, etc.).
- Possibilty to manage the functions outside the admin application using a CLI tool or API calls.

### Very hard

- Add support for functions written in WASM using wazero. (not sure)
  - We might need to be able to manage the functions outside the admin.
  - Error handling needs to be abstracted.
  - We need to define the WASM function interface.
