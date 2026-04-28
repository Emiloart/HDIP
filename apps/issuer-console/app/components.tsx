"use client";

import type {
  CredentialRecord,
  IssuanceResponse,
} from "@hdip/api-client";
import Link from "next/link";
import { useCallback, useEffect, useMemo, useState, type FormEvent } from "react";

import { issuerApi } from "../lib/api";
import {
  availableStatusActions,
  createIssuanceRequest,
  credentialStatusUpdateRequest,
  defaultCreateCredentialFormState,
  formatDateTime,
  idempotencyKey,
  mergeRecentCredentials,
  recentCredentialStorageKey,
  serviceErrorMessage,
  toRecentCredential,
  type CreateCredentialFormState,
  type RecentCredential,
  type StatusAction,
} from "../lib/issuer-console-state";

type FieldName = keyof CreateCredentialFormState;

function updateRecentCredential(next: RecentCredential) {
  const existing = readRecentCredentials();
  window.localStorage.setItem(
    recentCredentialStorageKey,
    JSON.stringify(mergeRecentCredentials(existing, next)),
  );
}

function readRecentCredentials(): RecentCredential[] {
  try {
    const raw = window.localStorage.getItem(recentCredentialStorageKey);
    if (raw === null) {
      return [];
    }

    const parsed = JSON.parse(raw) as RecentCredential[];
    if (!Array.isArray(parsed)) {
      return [];
    }

    return parsed.filter((item) => typeof item.credentialId === "string");
  } catch {
    return [];
  }
}

export function CreateCredentialWorkflow() {
  const [form, setForm] = useState(defaultCreateCredentialFormState);
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [created, setCreated] = useState<IssuanceResponse | null>(null);

  function setField(field: FieldName, value: string) {
    setForm((current) => ({ ...current, [field]: value }));
  }

  async function onSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault();
    setIsSubmitting(true);
    setError(null);
    setCreated(null);

    try {
      const response = await issuerApi.issueCredential(createIssuanceRequest(form), {
        idempotencyKey: idempotencyKey("issuer-create"),
      });
      setCreated(response);
      updateRecentCredential({
        credentialId: response.credentialId,
        status: response.status,
        expiresAt: response.expiresAt,
      });
    } catch (caughtError) {
      setError(serviceErrorMessage(caughtError));
    } finally {
      setIsSubmitting(false);
    }
  }

  return (
    <section className="console-panel" aria-busy={isSubmitting}>
      <div className="panel-heading">
        <p className="eyebrow">Issuer operations</p>
        <h1>Create credential</h1>
      </div>
      <form className="credential-form" onSubmit={onSubmit}>
        <label>
          Template ID
          <input value={form.templateId} onChange={(event) => setField("templateId", event.target.value)} required />
        </label>
        <label>
          Subject reference
          <input value={form.subjectReference} onChange={(event) => setField("subjectReference", event.target.value)} required />
        </label>
        <label>
          Full legal name
          <input value={form.fullLegalName} onChange={(event) => setField("fullLegalName", event.target.value)} required />
        </label>
        <label>
          Date of birth
          <input type="date" value={form.dateOfBirth} onChange={(event) => setField("dateOfBirth", event.target.value)} required />
        </label>
        <label>
          Country of residence
          <input maxLength={2} value={form.countryOfResidence} onChange={(event) => setField("countryOfResidence", event.target.value)} required />
        </label>
        <label>
          Document country
          <input maxLength={2} value={form.documentCountry} onChange={(event) => setField("documentCountry", event.target.value)} required />
        </label>
        <label>
          KYC level
          <input value={form.kycLevel} onChange={(event) => setField("kycLevel", event.target.value)} required />
        </label>
        <label>
          Verified at
          <input type="datetime-local" value={form.verifiedAt} onChange={(event) => setField("verifiedAt", event.target.value)} required />
        </label>
        <label>
          Expires at
          <input type="datetime-local" value={form.expiresAt} onChange={(event) => setField("expiresAt", event.target.value)} required />
        </label>
        <button type="submit" disabled={isSubmitting}>
          {isSubmitting ? "Creating..." : "Create credential"}
        </button>
      </form>
      {error !== null ? <p className="error-state" role="alert">{error}</p> : null}
      {created !== null ? <CredentialCreatedSummary response={created} /> : null}
    </section>
  );
}

function CredentialCreatedSummary(props: { response: IssuanceResponse }) {
  const artifact = JSON.stringify(props.response.credentialArtifact);

  async function copyArtifact() {
    await navigator.clipboard.writeText(artifact);
  }

  return (
    <section className="result-panel" aria-label="Created credential">
      <h2>Credential created</h2>
      <dl className="detail-grid">
        <div>
          <dt>Credential ID</dt>
          <dd>{props.response.credentialId}</dd>
        </div>
        <div>
          <dt>Status</dt>
          <dd>{props.response.status}</dd>
        </div>
        <div>
          <dt>Expires</dt>
          <dd>{formatDateTime(props.response.expiresAt)}</dd>
        </div>
        <div>
          <dt>Status reference</dt>
          <dd>{props.response.statusReference}</dd>
        </div>
      </dl>
      <textarea readOnly value={artifact} aria-label="Opaque credential artifact" />
      <div className="button-row">
        <button type="button" onClick={copyArtifact}>Copy artifact</button>
        <Link className="button-link" href={`/credentials/${encodeURIComponent(props.response.credentialId)}`}>
          Open detail
        </Link>
      </div>
    </section>
  );
}

export function CredentialLookupWorkflow() {
  const [credentialId, setCredentialId] = useState("");
  const [recentCredentials, setRecentCredentials] = useState<RecentCredential[]>([]);

  useEffect(() => {
    const timer = window.setTimeout(() => {
      setRecentCredentials(readRecentCredentials());
    }, 0);

    return () => window.clearTimeout(timer);
  }, []);

  const normalizedCredentialId = credentialId.trim();

  return (
    <section className="console-panel">
      <div className="panel-heading">
        <p className="eyebrow">Credential operations</p>
        <h1>Credentials</h1>
      </div>
      <form className="lookup-form" action={normalizedCredentialId === "" ? "/credentials" : `/credentials/${encodeURIComponent(normalizedCredentialId)}`}>
        <label>
          Credential ID
          <input
            value={credentialId}
            onChange={(event) => setCredentialId(event.target.value)}
            placeholder="cred_hdip_passport_basic_001"
          />
        </label>
        <button type="submit" disabled={normalizedCredentialId === ""}>Search</button>
      </form>
      <section className="table-panel">
        <h2>Recent credentials</h2>
        {recentCredentials.length === 0 ? (
          <p className="empty-state">No credentials created in this browser session.</p>
        ) : (
          <table>
            <thead>
              <tr>
                <th>Credential ID</th>
                <th>Status</th>
                <th>Expires</th>
              </tr>
            </thead>
            <tbody>
              {recentCredentials.map((credential) => (
                <tr key={credential.credentialId}>
                  <td>
                    <Link href={`/credentials/${encodeURIComponent(credential.credentialId)}`}>
                      {credential.credentialId}
                    </Link>
                  </td>
                  <td>{credential.status}</td>
                  <td>{formatDateTime(credential.expiresAt)}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>
    </section>
  );
}

export function CredentialDetailWorkflow(props: { credentialId: string }) {
  const [record, setRecord] = useState<CredentialRecord | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [pendingAction, setPendingAction] = useState<StatusAction | null>(null);
  const [supersededByCredentialId, setSupersededByCredentialId] = useState("");
  const [isUpdating, setIsUpdating] = useState(false);

  const loadCredential = useCallback(async () => {
    setIsLoading(true);
    setError(null);

    try {
      const loadedRecord = await issuerApi.credential(props.credentialId);
      setRecord(loadedRecord);
      updateRecentCredential(toRecentCredential(loadedRecord));
    } catch (caughtError) {
      setError(serviceErrorMessage(caughtError));
    } finally {
      setIsLoading(false);
    }
  }, [props.credentialId]);

  useEffect(() => {
    void loadCredential();
  }, [loadCredential]);

  const statusActions = useMemo(() => (record === null ? [] : availableStatusActions(record)), [record]);

  async function confirmStatusUpdate() {
    if (pendingAction === null || record === null) {
      return;
    }

    setIsUpdating(true);
    setError(null);

    try {
      await issuerApi.updateCredentialStatus(
        record.credentialId,
        credentialStatusUpdateRequest(pendingAction, supersededByCredentialId),
        { idempotencyKey: idempotencyKey(`issuer-status-${pendingAction}`) },
      );
      setPendingAction(null);
      setSupersededByCredentialId("");
      await loadCredential();
    } catch (caughtError) {
      setError(serviceErrorMessage(caughtError));
    } finally {
      setIsUpdating(false);
    }
  }

  if (isLoading) {
    return <p className="loading-state">Loading credential...</p>;
  }

  return (
    <section className="console-panel" aria-busy={isUpdating}>
      <div className="panel-heading">
        <p className="eyebrow">Credential detail</p>
        <h1>{props.credentialId}</h1>
      </div>
      {error !== null ? <p className="error-state" role="alert">{error}</p> : null}
      {record !== null ? (
        <>
          <CredentialDetail record={record} />
          <div className="button-row">
            {statusActions.includes("revoked") ? (
              <button type="button" onClick={() => setPendingAction("revoked")}>Revoke</button>
            ) : null}
            {statusActions.includes("superseded") ? (
              <button type="button" onClick={() => setPendingAction("superseded")}>Supersede</button>
            ) : null}
            <button type="button" onClick={() => void loadCredential()}>Refresh</button>
          </div>
          {pendingAction !== null ? (
            <section className="modal-panel" role="dialog" aria-modal="true" aria-label="Confirm status update">
              <h2>Confirm {pendingAction}</h2>
              <p>This will move credential {record.credentialId} to {pendingAction}.</p>
              {pendingAction === "superseded" ? (
                <label>
                  Superseding credential ID
                  <input
                    value={supersededByCredentialId}
                    onChange={(event) => setSupersededByCredentialId(event.target.value)}
                    placeholder="cred_hdip_passport_basic_002"
                  />
                </label>
              ) : null}
              <div className="button-row">
                <button
                  type="button"
                  onClick={confirmStatusUpdate}
                  disabled={isUpdating || (pendingAction === "superseded" && supersededByCredentialId.trim() === "")}
                >
                  {isUpdating ? "Updating..." : "Confirm"}
                </button>
                <button type="button" onClick={() => setPendingAction(null)} disabled={isUpdating}>
                  Cancel
                </button>
              </div>
            </section>
          ) : null}
        </>
      ) : null}
    </section>
  );
}

function CredentialDetail(props: { record: CredentialRecord }) {
  const artifact = props.record.credentialArtifact === undefined
    ? null
    : JSON.stringify(props.record.credentialArtifact);

  async function copyArtifact() {
    if (artifact !== null) {
      await navigator.clipboard.writeText(artifact);
    }
  }

  return (
    <section className="result-panel">
      <dl className="detail-grid">
        <div>
          <dt>Credential ID</dt>
          <dd>{props.record.credentialId}</dd>
        </div>
        <div>
          <dt>Status</dt>
          <dd>{props.record.status}</dd>
        </div>
        <div>
          <dt>Issued</dt>
          <dd>{formatDateTime(props.record.issuedAt)}</dd>
        </div>
        <div>
          <dt>Expires</dt>
          <dd>{formatDateTime(props.record.expiresAt)}</dd>
        </div>
        <div>
          <dt>Status updated</dt>
          <dd>{formatDateTime(props.record.statusUpdatedAt)}</dd>
        </div>
        <div>
          <dt>Template</dt>
          <dd>{props.record.templateId}</dd>
        </div>
        <div>
          <dt>Subject reference</dt>
          <dd>{props.record.subjectReference}</dd>
        </div>
        <div>
          <dt>Artifact digest</dt>
          <dd>{props.record.artifactDigest}</dd>
        </div>
      </dl>
      {artifact !== null ? (
        <>
          <textarea readOnly value={artifact} aria-label="Opaque credential artifact" />
          <button type="button" onClick={copyArtifact}>Copy artifact</button>
        </>
      ) : null}
    </section>
  );
}
