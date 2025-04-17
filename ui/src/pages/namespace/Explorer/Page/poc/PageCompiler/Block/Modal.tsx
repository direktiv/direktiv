import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogTrigger,
} from "~/design/Dialog";

import { Block } from ".";
import { BlockWrapper } from "./utils/BlockWrapper";
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
 */
export const Modal = ({ blocks, trigger }: ModalType) => (
  <BlockWrapper>
    <Dialog>
      <DialogTrigger>Open</DialogTrigger>
      <DialogContent>
        <BlocksWrapper>
          {blocks.map((block, index) => (
            <Block key={index} block={block} />
          ))}
        </BlocksWrapper>
        <DialogFooter>
          <DialogClose asChild>
            <Button variant="ghost">Cancel</Button>
          </DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  </BlockWrapper>
);
