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

/**
 * example
  {
      "code": 406,
      "message": "invalid json data: invalid character '}' looking for beginning of object key string"
  }
 */
export const JqQueryErrorSchema = z.object({
  code: z.number(),
  message: z.string(),
});

export type JqQueryResultType = z.infer<typeof JqQueryResult>;
