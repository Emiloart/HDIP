import { z } from "zod";

export const errorEnvelopeSchema = z.object({
  error: z.object({
    code: z.string().min(1),
    message: z.string().min(1),
    requestId: z.string().min(1).optional(),
  }),
});

export const healthResponseSchema = z.object({
  status: z.enum(["ok", "ready"]),
  service: z.string().min(1),
  version: z.string().min(1),
});

export type ErrorEnvelope = z.infer<typeof errorEnvelopeSchema>;
export type HealthResponse = z.infer<typeof healthResponseSchema>;
