import React, {
  MutableRefObject,
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";

import clsx from "clsx";
import { twMerge } from "tailwind-merge";

export const useCombinedRefs = (...refs: any[]): MutableRefObject<any> => {
  const targetRef = useRef<any>();

  useEffect(() => {
    refs.forEach((ref) => {
      if (!ref) return;
      if (typeof ref === "function") {
        ref(targetRef.current);
      } else {
        ref.current = targetRef.current;
      }
    });
  }, [refs]);
  return targetRef;
};

const FileInput = React.forwardRef<
  HTMLInputElement,
  React.InputHTMLAttributes<HTMLInputElement>
>(({ className, ...props }, ref) => {
  const inputRef = useRef<HTMLInputElement>(null);
  const combined = useCombinedRefs(ref, inputRef);

  const handleClick = useCallback(() => {
    combined.current?.click();
  }, [combined]);

  const [files, setFiles] = useState<FileList | null>(null);

  const title = useMemo(() => {
    if (!files) {
      return "No file chosen";
    } else if (files.length === 1) {
      return files[0]?.name;
    } else if (files.length > 1) {
      return `${files.length} files`;
    }
  }, [files]);

  return (
    <div
      className={clsx(
        "flex h-9 rounded-md bg-transparent text-sm focus:outline-none focus:ring-2 focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50",
        "border-gray-4 placeholder:text-gray-8 focus:ring-gray-4 focus:ring-offset-gray-1",
        "dark:border-gray-dark-4 dark:placeholder:text-gray-dark-8 dark:focus:ring-gray-dark-4 dark:focus:ring-offset-gray-dark-1"
      )}
    >
      <button
        onClick={handleClick}
        className="flex h-9 w-32 items-center justify-center rounded-l-lg  bg-black text-white active:scale-95 dark:bg-white dark:text-black"
      >
        choose file
      </button>
      <button
        onClick={handleClick}
        className="flex-1 rounded-r-lg border pl-2 text-left"
      >
        {title}
      </button>
      <input
        {...props}
        ref={combined}
        onChange={(e) => {
          setFiles(e.target.files);
          if (props.onChange) props.onChange(e);
        }}
        className={twMerge(clsx("hidden", className))}
        type="file"
      />
    </div>
  );
});

FileInput.displayName = "FileInput";

export default FileInput;
