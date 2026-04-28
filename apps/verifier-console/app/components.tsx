"use client";

import type { VerificationResult } from "@hdip/api-client";
import { useState, type FormEvent } from "react";

import { verifierApi } from "../lib/api";
import {
  createVerificationRequest,
  defaultVerifyCredentialFormState,
  formatDateTime,
  idempotencyKey,
  serviceErrorMessage,
  type VerifyCredentialFormState,
} from "../lib/verifier-console-state";

type FieldName = keyof VerifyCredentialFormState;

export function VerifyCredentialWorkflow() {
  const [form, setForm] = useState(defaultVerifyCredentialFormState);
  const [result, setResult] = useState<VerificationResult | null>(null);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  function setField(field: FieldName, value: string) {
    setForm((current) => ({ ...current, [field]: value }));
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsSubmitting(true);
    setError(null);
    setResult(null);

    try {
      const response = await verifierApi.verifyCredential(createVerificationRequest(form), {
        idempotencyKey: idempotencyKey("verifier-create"),
      });
      setResult(response);
    } catch (caughtError) {
      setError(serviceErrorMessage(caughtError));
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <section className="verify-panel" aria-busy={isSubmitting}>
      <div className="panel-heading">
        <p className="eyebrow">Verifier operations</p>
        <h1>Verify credential</h1>
      </div>
      <form className="verify-form" onSubmit={onSubmit}>
        <label>
          Credential ID
          <input
            value={form.credentialId}
            onChange={(event) => setField("credentialId", event.target.value)}
            placeholder="cred_hdip_passport_basic_001"
          />
        </label>
        <label>
          Credential artifact
          <textarea
            value={form.credentialArtifact}
            onChange={(event) => setField("credentialArtifact", event.target.value)}
            placeholder='{"kind":"phase1_opaque_artifact","mediaType":"application/vnd.hdip.phase1-opaque-artifact","value":"opaque-artifact:v1:..."}'
            required
          />
        </label>
        <button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "Verifying..." : "Verify"}
        </button>
      </form>
      {error !== null ? <p className="error-state" role="alert">{error}</p> : null}
      {result !== null ? <VerificationResultView result={result} /> : null}
    </section>
  );
}

function VerificationResultView(props: { result: VerificationResult }) {
  return (
    <section className={`result-panel decision-${props.result.decision}`} aria-label="Verification result">
      <h2>{props.result.decision}</h2>
      <dl className="result-grid">
        <div>
          <dt>Verification ID</dt>
          <dd>{props.result.verificationId}</dd>
        </div>
        <div>
          <dt>Evaluated</dt>
          <dd>{formatDateTime(props.result.evaluatedAt)}</dd>
        </div>
      </dl>
    </section>
  );
}
