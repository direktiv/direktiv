import { BlockPath, addSegmentsToPath } from "./utils/blockPath";
import {
  DialogClose,
  DialogContent,
  Dialog as DialogDesignComponent,
  DialogFooter,
  DialogTrigger,
} from "~/design/Dialog";

import { Block } from ".";
import { BlockList } from "./utils/BlockList";
import { Button } from "./Button";
import ButtonDesignComponent from "~/design/Button";
import { DialogType } from "../../schema/blocks/dialog";

/**
 *
 * TODO:
 * [] add title
 * [] add a concept for a submit and cancel as soon as the form is implemented
 * [] optionally there could always be an X icon at the top right corner
 */
type DialogProps = {
  blockProps: DialogType;
  blockPath: BlockPath;
};
export const Dialog = ({ blockProps, blockPath }: DialogProps) => {
  const { blocks, trigger } = blockProps;
  return (
    <DialogDesignComponent>
      <DialogTrigger asChild>
        <Button blockProps={trigger} />
      </DialogTrigger>
      <DialogContent>
        <BlockList>
          {blocks.map((block, index) => (
            <Block
              key={index}
              block={block}
              blockPath={addSegmentsToPath(blockPath, index)}
            />
          ))}
        </BlockList>
        <DialogFooter>
          <DialogClose asChild>
            <ButtonDesignComponent variant="ghost">
              Cancel
            </ButtonDesignComponent>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </DialogDesignComponent>
  );
};
