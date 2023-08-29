import {
  MirrorAuthType,
  MirrorFormSchema,
  MirrorKeepSSHKeysFormSchema,
  MirrorKeepTokenFormSchema,
  MirrorSshFormSchema,
  MirrorTokenFormSchema,
} from "../schema";

import { fileNameSchema } from "~/api/tree/schema/node";
import { z } from "zod";
import { zodResolver } from "@hookform/resolvers/zod";

type getResolverArgs = (input: {
  isMirror: boolean;
  isNew: boolean;
  authType: MirrorAuthType;
  keepCredentials: boolean;
  existingNamespaces: string[];
}) => ReturnType<typeof zodResolver>;

export const getResolver: getResolverArgs = ({
  isMirror,
  isNew,
  authType,
  keepCredentials,
  existingNamespaces,
}) => {
  const newNameSchema = fileNameSchema.and(
    z.string().refine((name) => !existingNamespaces.some((n) => n === name), {
      message: "The name already exists",
    })
  );

  const baseSchema = z.object({ name: isNew ? newNameSchema : z.string() });

  if (!isMirror) {
    return zodResolver(baseSchema);
  }
  if (keepCredentials && authType === "token") {
    return zodResolver(baseSchema.and(MirrorKeepTokenFormSchema));
  }
  if (keepCredentials && authType === "ssh") {
    return zodResolver(baseSchema.and(MirrorKeepSSHKeysFormSchema));
  }
  if (authType === "token") {
    return zodResolver(baseSchema.and(MirrorTokenFormSchema));
  }
  if (authType === "ssh") {
    return zodResolver(baseSchema.and(MirrorSshFormSchema));
  }
  return zodResolver(baseSchema.and(MirrorFormSchema));
};
