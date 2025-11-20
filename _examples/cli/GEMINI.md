# Gemini CLI Example

This example demonstrates how to use the `gemini` command-line interface to interact with the Gemini model running on a local server.

## Basic Usage

To run the `gemini` CLI and connect to a local server, set the `GOOGLE_GEMINI_BASE_URL` environment variable:

```sh
GOOGLE_GEMINI_BASE_URL=http://localhost:4567/ gemini
```

## Example Prompt

You can provide a prompt directly to the `gemini` CLI. For instance, to ask a question:

```sh
GOOGLE_GEMINI_BASE_URL=http://localhost:4567/ gemini "Explain the concept of recursion in programming."
```

## Expected Output

The CLI will then output the response from the Gemini model:

```text
Recursion in programming is a method where a function calls itself to solve a problem. Think of it like a set of Russian nesting dolls, where each doll contains a smaller version of itself.
... (rest of the explanation) ...
```
