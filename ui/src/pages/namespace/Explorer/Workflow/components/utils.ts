import { z } from "zod";

export const workflowInputSchema = z.string().refine((string) => {
  try {
    JSON.parse(string);
    return true;
  } catch (error) {
    return false;
  }
});
