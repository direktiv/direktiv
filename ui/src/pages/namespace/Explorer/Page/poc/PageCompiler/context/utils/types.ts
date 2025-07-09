import {
  AllBlocksType,
  FormBlocks,
  InlineBlocksType,
} from "../../../schema/blocks";

import { BlockEditFormProps } from "../../../BlockEditor";
import { BlockPathType } from "../../Block";
import { DirektivPagesType } from "../../../schema";
import { LucideIcon } from "lucide-react";

// common types for block type configuration
type BlockTypeConfigBase = {
  label: string;
  icon: LucideIcon;
  allow: (page: DirektivPagesType, path: BlockPathType) => boolean;
};

// inline blocks don't have a form component
type BlockTypeConfigWithoutForm = {
  [K in InlineBlocksType]: {
    type: K;
    formComponent?: never;
    defaultValues: Extract<AllBlocksType, { type: K }>;
  };
}[InlineBlocksType];

// blocks that require a form, must have a form component that implments a form for that very block type
type BlockTypeConfigWithForm = {
  [K in FormBlocks as K["type"]]: {
    type: K["type"];
    formComponent: React.ComponentType<BlockEditFormProps<K>>;
    defaultValues: K;
  };
}[FormBlocks["type"]];

export type BlockTypeConfig = BlockTypeConfigBase &
  (BlockTypeConfigWithForm | BlockTypeConfigWithoutForm);
