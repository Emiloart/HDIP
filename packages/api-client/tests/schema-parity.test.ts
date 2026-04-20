import fs from "node:fs";

import { describe, expect, it } from "vitest";

import {
  credentialTemplateMetadataSchema,
  errorEnvelopeSchema,
  healthResponseSchema,
  issuerProfileSchema,
  verifierPolicyRequestSchema,
  verifierResultSchema,
} from "../src/schemas";

type ContractName =
  | "errorEnvelope"
  | "healthResponse"
  | "issuerProfile"
  | "verifierPolicyRequest"
  | "verifierResult"
  | "credentialTemplateMetadata";

type ManifestEntry = {
  contract: ContractName;
  schema: string;
  fixture: string;
  valid: boolean;
};

type Manifest = {
  examples: ManifestEntry[];
};

const manifestPath = new URL("../../../schemas/examples/manifest.json", import.meta.url);
const exampleRoot = new URL("../../../schemas/examples/", import.meta.url);
const manifest = JSON.parse(fs.readFileSync(manifestPath, "utf8")) as Manifest;

const contractSchemas = {
  errorEnvelope: errorEnvelopeSchema,
  healthResponse: healthResponseSchema,
  issuerProfile: issuerProfileSchema,
  verifierPolicyRequest: verifierPolicyRequestSchema,
  verifierResult: verifierResultSchema,
  credentialTemplateMetadata: credentialTemplateMetadataSchema,
} as const;

describe("schema parity", () => {
  for (const entry of manifest.examples) {
    it(`matches ${entry.contract} example ${entry.fixture}`, () => {
      const fixturePath = new URL(entry.fixture, exampleRoot);
      const payload = JSON.parse(fs.readFileSync(fixturePath, "utf8")) as unknown;
      const result = contractSchemas[entry.contract].safeParse(payload);
      expect(result.success).toBe(entry.valid);
    });
  }
});
