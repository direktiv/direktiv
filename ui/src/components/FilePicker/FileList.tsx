import { FC, Fragment } from "react";
import {
  FilepickerClose,
  FilepickerListItem,
  FilepickerSelectButton,
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
  selectableFolders: boolean;
};

export const FileList: FC<FileListProps> = ({
  files,
  selectable,
  setPath,
  setInputValue,
  onChange,
  selectableFolders,
}) => (
  <>
    {files.map((file, index) => {
      const isSelectable = selectable?.(file) ?? true;
      const isLastListItem = index === files.length - 1;
      const filename = getFilenameFromPath(file.path);
      const parent = getParentFromPath(file.path);
      return (
        <Fragment key={filename}>
          {selectableFolders ? (
            <div>
              {isSelectable ? (
                <div className="group flex h-auto w-full cursor-pointer items-center justify-between hover:bg-gray-3">
                  <div
                    onClick={() => {
                      setPath(file.path);
                    }}
                    className="pr-2 text-gray-11 hover:bg-gray-3 hover:underline focus:bg-transparent focus:ring-0 focus:ring-transparent focus:ring-offset-0 dark:text-gray-dark-11 dark:hover:bg-gray-dark-3 dark:focus:bg-transparent"
                  >
                    <FilepickerListItem icon={fileTypeToIcon(file.type)}>
                      {filename}
                    </FilepickerListItem>
                  </div>
                  <div className="h-auto px-4 py-2 opacity-0 group-hover:opacity-100">
                    <FilepickerSelectButton
                      onClick={() => {
                        setPath(parent);
                        setInputValue(file.path);
                        onChange?.(file.path);
                      }}
                    >
                      Select
                    </FilepickerSelectButton>
                  </div>
                </div>
              ) : (
                <div className="flex h-auto w-full cursor-not-allowed items-center justify-between opacity-70 hover:bg-gray-3 dark:hover:bg-gray-dark-3">
                  <div
                    className={twMergeClsx(
                      "h-auto w-full text-gray-11 dark:text-gray-dark-11",
                      !isSelectable && "cursor-not-allowed opacity-70"
                    )}
                  >
                    <FilepickerListItem icon={fileTypeToIcon(file.type)}>
                      {filename}
                    </FilepickerListItem>
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div>
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
            </div>
          )}

          {!isLastListItem && <FilepickerSeparator />}
        </Fragment>
      );
    })}
  </>
);
