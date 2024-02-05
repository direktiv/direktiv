import { z } from "zod";

/**
 * example
  {
    "type":  "secret",
    "id":  "ACCESS_KEY",
    "issue":  "secret 'ACCESS_KEY' has not been initialized",
    "level":  "critical"
  }
 */
const LintingIssueSchema = z.object({
  type: z.enum(["secret"]),
  id: z.string(),
  issue: z.string(),
  level: z.string(),
});

/**
 * example
  {
    "namespace":  {
      "createdAt":  "2023-10-04T07:43:17.082556Z",
      "updatedAt":  "2023-10-04T07:43:17.082556Z",
      "name":  "dir-672",
      "notes":  {
        "commit_hash":  "3c6d83c6e852c1197e84e2fa474d7f70f46065e3",
        "ref":  "main",
        "url":  "https://github.com/direktiv/direktiv-examples.git"
      }
    },
    "issues":  []
  }
 */
export const LintSchema = z.object({
  issues: z.array(LintingIssueSchema),
});

export type LintSchemaType = z.infer<typeof LintSchema>;
