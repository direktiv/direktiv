import { FC, Fragment } from "react";
import {
  FilepickerClose,
  FilepickerListItem,
  FilepickerSeparator,
} from "~/design/Filepicker";
import {
  fileTypeToIcon,
  getFilenameFromPath,
  getParentFromPath,
} from "~/api/files/utils";

import { BaseFileSchemaType } from "~/api/files/schema";
import { twMergeClsx } from "~/util/helpers";

export type FileListProps = {
  files: BaseFileSchemaType[];
  selectable: ((file: BaseFileSchemaType) => boolean) | undefined;
  setPath: (path: string) => void;
  setInputValue: (value: string) => void;
  onChange: (path: string) => void;
};

export const FileList: FC<FileListProps> = ({
  files,
  selectable,
  setPath,
  setInputValue,
  onChange,
}) => (
  <>
    {files.map((file, index) => {
      const isSelectable = selectable?.(file) ?? true;
      const isLastListItem = index === files.length - 1;
      const filename = getFilenameFromPath(file.path);
      const parent = getParentFromPath(file.path);
      return (
        <Fragment key={filename}>
          {file.type === "directory" ? (
            <div
              onClick={() => {
                setPath(file.path);
              }}
              className="h-auto w-full cursor-pointer text-gray-11 hover:underline focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:focus:bg-transparent"
            >
              <FilepickerListItem icon={fileTypeToIcon(file.type)}>
                {filename}
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
                setPath(parent);
                setInputValue(file.path);
                onChange?.(file.path);
              }}
            >
              <FilepickerListItem icon={fileTypeToIcon(file.type)}>
                {filename}
              </FilepickerListItem>
            </FilepickerClose>
          )}
          {!isLastListItem && <FilepickerSeparator />}
        </Fragment>
      );
    })}
  </>
);
