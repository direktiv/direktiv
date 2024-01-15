import { envVariablesSchema } from "./schema";
import { z } from "zod";

/**
 * we are typing all env variables as optional (using Partial), no matter
 * what the schema says. We are optimistically defining all env variables
 * globally here, without really knowing if the zod schema has been applied.
 *
 * E.g. when a file will be executed in a vite context, the env variables
 * will be available, but when a file is executed in a node context (like
 * a api factory function), the env variables will not be available.
 *
 * typing them as optional will force us to cover both cases.
 */
type EnvVariableTypes = {
  VITE: Partial<z.infer<typeof envVariablesSchema>> | undefined;
};

declare global {
  // eslint-disable-next-line @typescript-eslint/no-namespace
  namespace NodeJS {
    // eslint-disable-next-line @typescript-eslint/no-empty-interface -- {} is needed to also keep the existing NODE_ENV type (like e.g process.env.TZ)
    interface ProcessEnv extends EnvVariableTypes {}
  }
}
