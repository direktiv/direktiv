import { FC, Fragment } from "react";
import {
  FilepickerClose,
  FilepickerListItem,
  FilepickerSeparator,
} from "~/design/Filepicker";

import { NodeSchemaType } from "~/api/tree/schema/node";
import { fileTypeToIcon } from "~/api/tree/utils";
import { twMergeClsx } from "~/util/helpers";

export type FileListProps = {
  results: NodeSchemaType[];
  selectable: ((node: NodeSchemaType) => boolean) | undefined;
  setPath: (path: string) => void;
  setInputValue: (value: string) => void;
  onChange: (path: string) => void;
};

export const FileList: FC<FileListProps> = ({
  results,
  selectable,
  setPath,
  setInputValue,
  onChange,
}) => (
  <>
    {results.map((file, index) => {
      const isSelectable = selectable?.(file) ?? true;
      const isLastListItem = index === results.length - 1;
      return (
        <Fragment key={file.name}>
          {file.type === "directory" ? (
            <div
              onClick={() => {
                setPath(file.path);
              }}
              className="h-auto w-full cursor-pointer text-gray-11 hover:underline focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:focus:bg-transparent"
            >
              <FilepickerListItem icon={fileTypeToIcon(file.type)}>
                {file.name}
              </FilepickerListItem>
            </div>
          ) : (
            <FilepickerClose
              className={twMergeClsx(
                "h-auto w-full text-gray-11 hover:underline dark:text-gray-dark-11",
                !isSelectable && "cursor-not-allowed opacity-70"
              )}
              disabled={!isSelectable}
              onClick={() => {
                setPath(file.parent);
                setInputValue(file.path);
                onChange?.(file.path);
              }}
            >
              <FilepickerListItem icon={fileTypeToIcon(file.type)}>
                {file.name}
              </FilepickerListItem>
            </FilepickerClose>
          )}
          {!isLastListItem && <FilepickerSeparator />}
        </Fragment>
      );
    })}
  </>
);
