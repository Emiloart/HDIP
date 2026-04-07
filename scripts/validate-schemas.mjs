import fs from "node:fs/promises";
import path from "node:path";

import Ajv2020 from "ajv/dist/2020.js";
import addFormats from "ajv-formats";

const root = process.cwd();
const schemaRoot = path.join(root, "schemas", "json");

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

for (const file of schemaFiles) {
  const raw = await fs.readFile(file, "utf8");
  const schema = JSON.parse(raw);
  ajv.compile(schema);
}

console.log(`Validated ${schemaFiles.length} schema files.`);
