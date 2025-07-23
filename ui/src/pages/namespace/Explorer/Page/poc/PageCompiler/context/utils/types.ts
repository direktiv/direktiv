import { FormBlockType, NoFormBlockType } from "../../../schema/blocks";

import { BlockEditFormProps } from "../../../BlockEditor";
import { BlockPathType } from "../../Block";
import { ComponentType } from "react";
import { DirektivPagesType } from "../../../schema";
import { LucideIcon } from "lucide-react";

type ConfigBase = {
  label: string;
  icon: LucideIcon;
  allow: (page: DirektivPagesType, path: BlockPathType) => boolean;
};

// inline blocks don't have a form component
type WithForm = {
  [K in NoFormBlockType as K["type"]]: {
    type: K["type"];
    formComponent?: never;
    defaultValues: K;
  };
}[NoFormBlockType["type"]];

// BlockTypeConfigWithForm must have a form component that implements a form for that very block type
type WithoutForm = {
  [K in FormBlockType as K["type"]]: {
    type: K["type"];
    formComponent: ComponentType<BlockEditFormProps<K>>;
    defaultValues: K;
  };
}[FormBlockType["type"]];

export type BlockTypeConfig = ConfigBase & (WithoutForm | WithForm);
