import { FormBlocks, InlineBlocks } from "../../../schema/blocks";

import { BlockEditFormProps } from "../../../BlockEditor";
import { BlockPathType } from "../../Block";
import { ComponentType } from "react";
import { DirektivPagesType } from "../../../schema";
import { LucideIcon } from "lucide-react";

type BlockTypeConfigBase = {
  label: string;
  icon: LucideIcon;
  allow: (page: DirektivPagesType, path: BlockPathType) => boolean;
};

// inline blocks don't have a form component
type BlockTypeConfigWithoutForm = {
  [K in InlineBlocks as K["type"]]: {
    type: K["type"];
    formComponent?: never;
    defaultValues: K;
  };
}[InlineBlocks["type"]];

// BlockTypeConfigWithForm must have a form component that implements a form for that very block type
type BlockTypeConfigWithForm = {
  [K in FormBlocks as K["type"]]: {
    type: K["type"];
    formComponent: ComponentType<BlockEditFormProps<K>>;
    defaultValues: K;
  };
}[FormBlocks["type"]];

export type BlockTypeConfig = BlockTypeConfigBase &
  (BlockTypeConfigWithForm | BlockTypeConfigWithoutForm);
