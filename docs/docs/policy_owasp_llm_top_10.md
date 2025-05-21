# OWASP LLM Top 10 Policies (Examples)

This document describes example policies based on the [OWASP Top 10 for Large Language Model Applications (v1.1)](https://owasp.org/www-project-top-10-for-large-language-model-applications/). These policies are designed to be used with the Intercept engine to scan prompts or related configuration files for potential LLM-specific vulnerabilities.

**Policy File:** `playground/policies/owasp_llm_top_10.yaml`
**Target Path for Examples:** `playground/targets/prompts/`

## Implemented Policies

The following provides a brief overview of the implemented example policies.

### 1. LLM01: Prompt Injection (`LLM01-PROMPT-INJECTION-001`)

-   **Name:** Detect Common Prompt Injection Keywords
-   **Description:** Identifies common keywords and phrases associated with prompt injection attempts. These patterns may indicate an attempt to manipulate the LLM (e.g., "ignore previous instructions", "act as if you were").
-   **Severity:** Medium (Score: 7)
-   **Detection Method:** Regex scan for known injection phrases in `.txt` files.
-   **Example Triggers:**
    -   `ignore your previous instructions`
    -   `act as if you were an administrator`
    -   `reveal your prompt`
-   **Example Files:**
    -   `playground/targets/prompts/prompt_injection_example.txt` (should trigger)
    -   `playground/targets/prompts/clean_prompt_example.txt` (should not trigger)

### 2. LLM06: Sensitive Information Disclosure in Prompts

This category includes policies to detect various types of sensitive data that should not be present in prompts sent to LLMs.

#### a. Credit Card Numbers (`LLM06-SENSITIVE-DATA-CREDIT-CARD-001`)

-   **Name:** Detect Potential Credit Card Numbers in Prompts
-   **Description:** Identifies patterns resembling credit card numbers.
-   **Severity:** High (Score: 9)
-   **Detection Method:** Regex scan for common credit card number formats.
-   **Example Files:**
    -   `playground/targets/prompts/prompt_with_secrets.txt` (should trigger)

#### b. AWS Access Keys (`LLM06-SENSITIVE-DATA-AWS-KEYS-001`)

-   **Name:** Detect Potential AWS Access Keys in Prompts
-   **Description:** Identifies patterns resembling AWS Access Key IDs.
-   **Severity:** Medium (Score: 8)
-   **Detection Method:** Regex scan for AWS Access Key ID patterns.
-   **Example Files:**
    -   `playground/targets/prompts/prompt_with_secrets.txt` (should trigger)

#### c. Private Key Headers (`LLM06-SENSITIVE-DATA-PRIVATE-KEY-001`)

-   **Name:** Detect Potential Private Key Headers in Prompts
-   **Description:** Identifies common private key block headers (e.g., RSA, EC, OPENSSH).
-   **Severity:** High (Score: 9)
-   **Detection Method:** Regex scan for private key header patterns.
-   **Example Files:**
    -   `playground/targets/prompts/prompt_with_secrets.txt` (should trigger)

-   **Shared Clean Example File (for all LLM06 policies):**
    -   `playground/targets/prompts/prompt_without_secrets.txt` (should not trigger any LLM06 policies)


### 3. LLM04: Model Denial of Service (Prompt Length) (`LLM04-DOS-PROMPT-LENGTH-001`)

-   **Name:** Detect Overly Long Prompts
-   **Description:** Identifies prompts that exceed a defined character length (currently >4000 characters), which could be a factor in Model Denial of Service or hitting token limits.
-   **Severity:** Low (Score: 3)
-   **Detection Method:** Regex scan to check the total length of the file content.
-   **Example Files:**
    -   `playground/targets/prompts/long_prompt_example.txt` (should trigger)
    -   `playground/targets/prompts/normal_prompt_example.txt` (should not trigger)

## Usage

These policies can be used with tools like `intercept audit` by targeting the `playground/policies/owasp_llm_top_10.yaml` file and the relevant prompt files. For example:

```bash
# (Assuming intercept is in PATH and you are in the root of the repository)
intercept audit -p playground/policies/owasp_llm_top_10.yaml playground/targets/prompts/prompt_injection_example.txt
intercept audit -p playground/policies/owasp_llm_top_10.yaml playground/targets/prompts/prompt_with_secrets.txt
intercept audit -p playground/policies/owasp_llm_top_10.yaml playground/targets/prompts/long_prompt_example.txt
```

## Disclaimer

These policies are examples and provide a basic level of detection. They are not exhaustive and may produce false positives or false negatives. Effective LLM security requires a multi-layered approach, including input validation, output sanitization, content filtering, context-aware monitoring, and robust access controls. Always adapt and enhance policies based on your specific LLM usage and risk environment.
```
