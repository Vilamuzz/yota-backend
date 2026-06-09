# Skill: State, Logging, & Error Handling Rules

This document defines the strict patterns for handling execution anomalies, structuring logging outputs, and formatting unified API responses within the `vilamuzz-yota-backend` workspace. AI tools must strictly mirror these behaviors to maintain operational predictability.

---

## 1. Unified Response Envelope (`pkg.NewResponse`)

We enforce a standardized JSON envelope across the entire application using the `pkg.NewResponse` factory.

### Structure Protocol

The response contract follows this exact signature:

```go
pkg.NewResponse(statusCode int, message string, data any, meta any)
```

---

## 2. GORM Error Identification & Layer Mapping

We do not bubble up naked database driver errors to the transport layer. The Service Layer is explicitly responsible for catching GORM database primitives, evaluating them, and mapping them into human-readable `pkg.Response` structures.

### Not Found vs. Server Errors

- **`gorm.ErrRecordNotFound`**: Must be explicitly checked using `errors.Is(err, gorm.ErrRecordNotFound)` and converted to an `http.StatusNotFound` (404) response with localized, clear messages.
- **Unexpected Failures**: Fatal database drops or connection errors must be treated as `http.StatusInternalServerError` (500) to hide internal query structures from the client.

```go
donationProgram, err := s.repo.FindOneDonationProgram(ctx, map[string]interface{}{"id": id})
if err != nil {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return pkg.NewResponse(http.StatusNotFound, "Donasi tidak ditemukan", nil, nil)
	}
	return pkg.NewResponse(http.StatusInternalServerError, "Terjadi kesalahan pada server", nil, nil)
}
```

---

## 3. Structural Log Capture Policy

When a state mutation or infrastructure command fails unexpectedly (e.g., failed DB updates, I/O errors), logs must be safely captured inside the Service layer before returning the response payload.

### Context-Rich Logging

Logs must include fields that pinpoint the failure context (component and relevant entity primary keys) using structural logging contexts with `logrus`:

```go
if err := s.repo.UpdateDonationProgram(ctx, id, updateData); err != nil {
	// Structural context field logging
	logrus.WithFields(logrus.Fields{
		"component":   "donation.service",
		"donation_id": id,
	}).WithError(err).Error("failed to update donation")

	return pkg.NewResponse(http.StatusInternalServerError, "Gagal memperbarui donasi", nil, nil)
}
```

---

## 4. ❌ FORBIDDEN AI ANTI-PATTERNS (DO NOT GENERATE)

- **Naked Error Returns**: Do not return an unmapped Go error primitive directly from a service method to a handler. Every business service outcome must return a correctly populated `pkg.Response` object.
- **Silent Database Failures**: Do not omit error checks on GORM transactions or updates. Every write execution (Create, Save, Update, Delete) must have an active conditional block ensuring failure capture.
- **HTTP Status Code Mismatches**: Do not return raw string messages with mismatched status code headers (e.g., sending `http.StatusOK` with a message body saying "Failed").
- **Log Masking**: Do not log errors using plain unformatted text like `log.Print(err)`. You must use `logrus.WithFields` to attach structural tracking properties.
