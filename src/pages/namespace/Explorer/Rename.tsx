import {
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../../../design/Dialog";
import { TextCursorInput, Trash } from "lucide-react";

import Button from "../../../design/Button";
import Input from "../../../design/Input";
import { NodeSchemaType } from "../../../api/tree/schema";
import { useRenameNode } from "../../../api/tree/mutate/renameNode";
import { useState } from "react";

const Rename = ({
  node,
  close,
}: {
  node: NodeSchemaType;
  close: () => void;
}) => {
  const [name, setName] = useState(node.name);
  const { mutate, isLoading } = useRenameNode({
    onSuccess: () => {
      close();
    },
  });

  return (
    <>
      <DialogHeader>
        <DialogTitle>
          <Trash /> Delete
        </DialogTitle>
      </DialogHeader>
      <div className="my-3">
        <Input value={name} onChange={(e) => setName(e.target.value)} />
      </div>
      <DialogFooter>
        <DialogClose asChild>
          <Button variant="ghost">Cancel</Button>
        </DialogClose>
        <Button
          onClick={() => {
            mutate({ node, newName: name });
          }}
          loading={isLoading}
        >
          {!isLoading && <TextCursorInput />}
          Rename
        </Button>
      </DialogFooter>
    </>
  );
};

export default Rename;
