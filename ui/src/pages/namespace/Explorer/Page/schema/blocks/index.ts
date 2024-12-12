import { AllBlocksType } from "./types";
import { Button } from "./button";
import { Form } from "./form";
import { Headline } from "./headline";
import { Modal } from "./modal";
import { QueryProvider } from "./queryProvider";
import { Text } from "./text";
import { TwoColumnsType } from "./twoColumns";
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

export const TwoColumns = z.object({
  type: z.literal("two-columns"),
  leftBlocks: z.array(AllBlocks),
  rightBlocks: z.array(AllBlocks),
}) satisfies z.ZodType<TwoColumnsType>;

export const TriggerBlocks = z.discriminatedUnion("type", [Button]);
