import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "../Dialog";
import { DroppableElement, Placeholder } from "./DroppableElement";
import { FC, PropsWithChildren, ReactNode, useState } from "react";
import { Image, List, TableIcon } from "lucide-react";
import { MainContent, MainTop } from "../Appshell";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "../Select";
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeaderCell,
  TableRow,
} from "../Table";

import Avatar from "../Avatar";
import Badge from "../Badge";
import Button from "../Button";
import { Card } from "../Card";
import { DndContext } from "./Context.tsx";
import { Draggable } from "./DraggableElement";
import Input from "../Input";
import { twMergeClsx } from "~/util/helpers";

const RenderLayout: FC<PropsWithChildren> = ({ children }) => (
  <div className={twMergeClsx("h-24 w-full bg-slate-100 p-4")}>{children}</div>
);

type person = {
  name?: string;
  email?: string;
  title?: string;
  role?: string;
};

type TableProps = {
  columns?: number;
  rows?: number;
  people?: person[];
};

const DefaultTable: FC<TableProps> = () => (
  <Table>
    <TableHead>
      <TableRow>
        <TableHeaderCell>...</TableHeaderCell>
        <TableHeaderCell>...</TableHeaderCell>
        <TableHeaderCell>...</TableHeaderCell>
        <TableHeaderCell>...</TableHeaderCell>
        <TableHeaderCell>
          <span className="sr-only">Edit</span>
        </TableHeaderCell>
      </TableRow>
    </TableHead>
    <TableBody>
      <TableRow>
        <TableCell>...</TableCell>
        <TableCell>...</TableCell>
        <TableCell>...</TableCell>
        <TableCell>...</TableCell>
      </TableRow>
    </TableBody>
  </Table>
);

const DefaultList: FC = () => (
  <ul className="list-disc">
    <li>...</li>
    <li>...</li>
    <li>...</li>
  </ul>
);

const getElementComponent = (element: string) => {
  switch (element) {
    case "image":
      return <Avatar />;
    case "table":
      return <DefaultTable />;
    case "list":
      return <DefaultList />;
    default:
      return <div></div>;
  }
};

export const DragAndDropEditor: FC = () => {
  const [component, setComponent] = useState<ReactNode>(<div></div>);

  const [dialogOpen, setDialogOpen] = useState<boolean>(false);
  const [header, setHeader] = useState<ReactNode>(<div></div>);
  const [footer, setFooter] = useState<ReactNode>(<div></div>);

  const onMove = (element: string, target: string) => {
    if (target) {
      const elementComponent = getElementComponent(element);
      setComponent(elementComponent);
    }
  };

  return (
    <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
      <DndContext onMove={onMove}>
        <div className="flex grow flex-col">
          <MainTop>
            <h3 className="flex justify-center space-x-3 text-xl font-bold text-gray-12 dark:text-gray-dark-12">
              Components
            </h3>
            <h3 className="flex justify-center space-x-3 text-xl font-bold text-gray-12 dark:text-gray-dark-12">
              Editor
            </h3>
            <h3 className="flex justify-center space-x-3 text-xl font-bold text-gray-12 dark:text-gray-dark-12">
              Preview
            </h3>
          </MainTop>
          <MainContent>
            <div className="flex flex-row">
              <div className="w-1/3 flex-col">
                <Card className="h-full bg-gray-1 p-4">
                  <Draggable name="image">
                    <Button asChild variant="outline" size="lg">
                      <div className="w-28 bg-white">
                        <Image size={16} />
                        Image
                      </div>
                    </Button>
                  </Draggable>
                  <Draggable name="table">
                    <Button asChild variant="outline" size="lg">
                      <div className="w-28 bg-white ">
                        <TableIcon size={16} />
                        Table
                      </div>
                    </Button>
                  </Draggable>
                  <Draggable name="list">
                    <Button asChild variant="outline" size="lg">
                      <div className="w-28 bg-white">
                        <List size={16} />
                        List
                      </div>
                    </Button>
                  </Draggable>
                </Card>
              </div>
              <div className="w-1/3 flex-col">
                <Card className="h-full bg-gray-1 p-4">
                  <Placeholder
                    onClick={() => setDialogOpen(true)}
                    name="Header"
                  />
                  <DroppableElement
                    onClick={() => setDialogOpen(true)}
                    position="1"
                  />

                  <Placeholder
                    onClick={() => setDialogOpen(true)}
                    name="Footer"
                  />
                </Card>
              </div>
              <div className="w-1/3 flex-col">
                <Card className="h-full bg-gray-1 p-4">
                  <RenderLayout>{header}</RenderLayout>
                  <RenderLayout>{component}</RenderLayout>
                  <RenderLayout>{footer}</RenderLayout>
                </Card>
              </div>
            </div>
          </MainContent>
        </div>
      </DndContext>

      <DialogContent>
        <DialogHeader>
          <DialogTitle>Edit Table</DialogTitle>
        </DialogHeader>
        <fieldset className="gap-5">
          <label className="w-[150px] text-right" htmlFor="name">
            API
          </label>
          <div className="w-full">
            <Select>
              <SelectTrigger variant="outline">
                <SelectValue placeholder="Select Data Source" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="userdirectories">
                  <Badge>GET</Badge> /userDirectories
                </SelectItem>
                <SelectItem value="userattributes">
                  <Badge>GET</Badge> /user/userAttributes
                </SelectItem>
              </SelectContent>
            </Select>
          </div>
        </fieldset>
        <fieldset className="gap-5">
          <label className="w-[150px] text-right" htmlFor="name">
            Add new items as
          </label>
          <div className="w-full">
            <Select>
              <SelectTrigger variant="outline">
                <SelectValue placeholder="Select table direction" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="rows">Rows</SelectItem>
                <SelectItem value="columns">Columns</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </fieldset>
        <fieldset className="gap-5">
          <label className="w-[150px] text-right" htmlFor="name">
            Header Cell
          </label>
          <Input
            id="name"
            data-testid="variable-name"
            placeholder="Please insert a name"
          />
        </fieldset>
        <fieldset className="gap-5">
          <label className="w-[150px] text-right" htmlFor="name">
            Table Cell
          </label>
          <div className="w-full">
            <Select>
              <SelectTrigger variant="outline">
                <SelectValue placeholder="Select Data Property" />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="attributeName">attributeName</SelectItem>
                <SelectItem value="createdAt">createdAt</SelectItem>
                <SelectItem value="updatedAt">updatedAt</SelectItem>
              </SelectContent>
            </Select>
          </div>
        </fieldset>

        <DialogFooter>
          <DialogClose asChild>
            <Button variant="outline" type="submit">
              Close
            </Button>
          </DialogClose>
          <Button variant="primary" type="submit">
            Save
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
};
