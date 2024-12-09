import { Button } from "./button";
import { Headline } from "./headline";
import { Modal } from "./modal";
import { Text } from "./text";
import { z } from "zod";

const AllBlocks = z.discriminatedUnion("type", [Headline, Button, Text, Modal]);

export const Block = {
  all: AllBlocks,
  trigger: z.discriminatedUnion("type", [Button]),
};

export type BlockType = {
  all: z.infer<typeof Block.all>;
  trigger: z.infer<typeof Block.trigger>;
};
