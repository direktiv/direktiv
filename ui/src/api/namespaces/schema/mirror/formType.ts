import { MirrorAuthTypeSchema } from ".";
import { z } from "zod";

const MirrorFormType = z.union([
  MirrorAuthTypeSchema,
  z.enum(["keep-ssh", "keep-token"]),
]);

export const getAuthTypeFromFormType = (formType: MirrorFormType) => {
  switch (formType) {
    case "keep-ssh":
    case "keep-token":
      return undefined;
    default:
      return formType;
  }
};

export type MirrorFormType = z.infer<typeof MirrorFormType>;
