Data Validator API

![Data Validator API - Thumbnail](assets/api-validation-thumbnail.png)

> Upwork portfolio thumbnail — Data validation API (email, CPF, phone, etc.)

The Data Validator API is a service designed to validate common data inputs, such as email, CPF, name, phone number, RG, CEP, and credit card number. It supports flexible input formats, sanitizes data to ensure consistency, and returns structured JSON responses.

Features

- Email Validation: Verifies email format and normalizes case.
- CPF Validation: Checks Brazilian CPF validity (including check digits) and caches result in Redis.
- Name Validation: Removes accents, strips invalid chars, normalizes spaces, and converts to uppercase.
- Phone Validation: Validates Brazilian phone numbers with DDD and mobile rules.
- RG Validation: Validates basic Brazilian RG format.
- CEP Validation: Validates CEP length and basic invalid values.
- Credit Card Validation: Uses Luhn algorithm and identifies card brand.

API Endpoint

GET /validate

This endpoint receives query parameters and validates exactly one parameter at a time.

Query Parameters

- email (optional) - Email address to validate.
- cpf (optional) - Brazilian CPF number to validate, accepts different formats.
- name (optional) - Name to validate and sanitize.
- telephone (optional) - Phone number to validate.
- phone (optional) - Alias for `telephone`.
- plastic (optional) - Credit card number to validate.
- rg (optional) - Brazilian RG to validate.
- cep (optional) - CEP (postal code) to validate.

Status Codes

- 200: Valid value.
- 400: Missing parameter or multiple parameters sent.
- 422: Parameter provided but failed validation.

Response Format

The API returns a JSON object with the validation result:

- status_code: HTTP status code.
- error_code: Optional machine-readable error code.
- parameter_key: Input parameter key.
- raw_parameter_value: Original input.
- parameter_value: Sanitized value.
- is_valid: Whether the value is valid.
- message: Human-readable message.
- request_id: Unique request identifier.
- timestamp: Request timestamp.
- execution_time_ms: Processing time.
- from_cache: Whether result came from cache (CPF only).

Examples

- Valid email
  - Request: `GET /validate?email=USER@Test.COM`
  - Response: `200` with `parameter_value: "user@test.com"` and `is_valid: true`.

- Invalid CPF
  - Request: `GET /validate?cpf=111.111.111-11`
  - Response: `422` with `error_code: "VALIDATION_FAILED"` and `is_valid: false`.

- Valid phone using alias
  - Request: `GET /validate?phone=+55 (11) 91234-5678`
  - Response: `200` with `parameter_key: "telephone"` and `parameter_value: "11912345678"`.

- Multiple parameters (not allowed)
  - Request: `GET /validate?email=a@a.com&cpf=52998224725`
  - Response: `400` with `error_code: "MULTIPLE_PARAMETERS"`.
