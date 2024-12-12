import { AllBlocksType } from "./types";
import { Button } from "./button";
import { Form } from "./form";
import { Headline } from "./headline";
import { Modal } from "./modal";
import { QueryProvider } from "./queryProvider";
import { Text } from "./text";
import { TwoColumns } from "./twoColumns";
import { z } from "zod";

export const AllBlocks: z.ZodType<AllBlocksType> = z.lazy(() =>
  z.discriminatedUnion("type", [
    Button,
    Form,
    Headline,
    Modal,
    QueryProvider,
    Text,
    TwoColumns,
  ])
);

export const TriggerBlocks = z.discriminatedUnion("type", [Button]);
