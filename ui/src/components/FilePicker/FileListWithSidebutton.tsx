import {
  CheckCircle2,
  Circle,
  CircleDashed,
  CircleDot,
  CircleOff,
} from "lucide-react";
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

export const FileListWithSidebutton: FC<FileListProps> = ({
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
          {isSelectable ? (
            <div className=" flex h-auto w-full cursor-pointer items-center justify-between hover:bg-gray-3">
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
              <div className="">
                <FilepickerClose
                  className={twMergeClsx(
                    "group h-auto px-4 py-2 text-gray-11 hover:bg-gray-3 dark:text-gray-dark-11  dark:hover:bg-gray-dark-3 "
                  )}
                  disabled={!isSelectable}
                  onClick={() => {
                    setPath(parent);
                    setInputValue(file.path);
                    onChange?.(file.path);
                  }}
                >
                  <div className=" h-4 w-4">
                    <CheckCircle2
                      aria-hidden="true"
                      className="invisible absolute h-4 w-4 text-gray-11 group-hover:visible"
                    />
                    <Circle
                      aria-hidden="true"
                      className=" absolute h-4 w-4 text-gray-11 "
                    />

                    {/* <CircleDot
                      aria-hidden="true"
                      className="invisible absolute h-4 w-4 text-gray-11 group-hover:visible"
                    /> */}
                  </div>
                </FilepickerClose>
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
          {!isLastListItem && <FilepickerSeparator />}
        </Fragment>
      );
    })}
  </>
);
