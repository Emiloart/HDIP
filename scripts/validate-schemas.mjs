import fs from "node:fs/promises";
import path from "node:path";

import Ajv2020 from "ajv/dist/2020.js";
import addFormats from "ajv-formats";

const root = process.cwd();
const schemaRoot = path.join(root, "schemas", "json");
const exampleRoot = path.join(root, "schemas", "examples");
const exampleManifestPath = path.join(exampleRoot, "manifest.json");

async function readJson(filePath) {
  return JSON.parse(await fs.readFile(filePath, "utf8"));
}

async function collectSchemaFiles(directory) {
  const entries = await fs.readdir(directory, { withFileTypes: true });
  const files = await Promise.all(
    entries.map(async (entry) => {
      const resolved = path.join(directory, entry.name);
      if (entry.isDirectory()) {
        return collectSchemaFiles(resolved);
      }

      if (entry.isFile() && entry.name.endsWith(".schema.json")) {
        return [resolved];
      }

      return [];
    }),
  );

  return files.flat();
}

const ajv = new Ajv2020({ strict: true, allErrors: true });
addFormats(ajv);

const schemaFiles = await collectSchemaFiles(schemaRoot);
if (schemaFiles.length === 0) {
  throw new Error("no schema files found");
}

const validators = new Map();

for (const file of schemaFiles) {
  const schema = await readJson(file);
  const relativeSchemaPath = path.relative(schemaRoot, file).split(path.sep).join("/");
  validators.set(relativeSchemaPath, ajv.compile(schema));
}

const manifest = await readJson(exampleManifestPath);
if (!Array.isArray(manifest.examples) || manifest.examples.length === 0) {
  throw new Error("no schema examples found in manifest");
}

for (const example of manifest.examples) {
  const validator = validators.get(example.schema);
  if (!validator) {
    throw new Error(`no validator found for schema ${example.schema}`);
  }

  const fixturePath = path.join(exampleRoot, example.fixture);
  const payload = await readJson(fixturePath);
  const isValid = validator(payload);

  if (isValid !== example.valid) {
    const details = validator.errors ? ajv.errorsText(validator.errors, { separator: "; " }) : "no details";
    throw new Error(
      `example ${example.fixture} expected valid=${example.valid} against ${example.schema}, got valid=${isValid}: ${details}`,
    );
  }
}

console.log(`Validated ${schemaFiles.length} schema files and ${manifest.examples.length} contract examples.`);
