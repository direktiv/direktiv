import { z } from "zod";
/**
 * example
  {
    "query":  ".foo[1]",
    "data":  "eyJmb28iOiBbeyJuYW1lIjoiSlNPTiIsICJnb29kIjp0cnVlfSwgeyJuYW1lIjoiWE1MIiwgImdvb2QiOmZhbHNlfV19",
    "results":  [
      "{\n  \"good\": false,\n  \"name\": \"XML\"\n}"
    ]
  }
 */

export const JqQueryResult = z.object({
  query: z.string(),
  data: z.string(),
  results: z.array(z.string()),
});

export type JqQueryResultType = z.infer<typeof JqQueryResult>;
