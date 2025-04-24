import { BlockPath, addSegmentsToPath } from "./utils/blockPath";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogTrigger,
} from "~/design/Dialog";

import { Block } from ".";
import { BlocksWrapper } from "./utils/BlocksWrapper";
import Button from "~/design/Button";
import { ModalType } from "../../schema/blocks/modal";

/**
 *
 * TODO:
 * [] rename Modal to Dialog
 * [] add title
 * [] add a concept for a submit and cancel as soon as the form is implemented
 * [] optionally there could always be an X icon at the top right corner
 * [] only render modal children when open
 *
 */

type ModalProps = {
  blockProps: ModalType;
  blockPath: BlockPath;
};
export const Modal = ({
  blockProps: { blocks, trigger },
  blockPath,
}: ModalProps) => (
  <Dialog>
    <DialogTrigger>Open</DialogTrigger>
    <DialogContent>
      <BlocksWrapper>
        {blocks.map((block, index) => (
          <Block
            key={index}
            block={block}
            blockPath={addSegmentsToPath(blockPath, index)}
          />
        ))}
      </BlocksWrapper>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">Cancel</Button>
        </DialogClose>
      </DialogFooter>
    </DialogContent>
  </Dialog>
);
