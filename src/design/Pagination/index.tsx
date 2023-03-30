import { ChevronLeftIcon, ChevronRightIcon } from "@heroicons/react/20/solid";

import { useState } from "react";

interface PaginationProps {
  total?: number;
  onChange: (index: number) => void;
}
export default function Pagination(props: PaginationProps) {
  const { total = 5, onChange } = props;
  const [current, setCurrent] = useState<number>(1);
  const handleChange = (index: number) => {
    setCurrent(index);
    onChange(index);
  };
  const handlePrev = () => {
    if (current > 1) {
      setCurrent(current - 1);
    }
  };
  const handleNext = () => {
    if (current < total) {
      setCurrent(current + 1);
    }
  };
  const inactiveClass =
    "relative inline-flex items-center px-4 py-2 text-sm font-semibold text-gray-900 ring-1 ring-inset ring-gray-3 hover:bg-gray-1 focus:z-20 focus:outline-offset-0 cursor-pointer";
  const activeClass =
    "relative z-10 inline-flex items-center bg-primary-600 px-4 py-2 text-sm font-semibold text-white focus:z-20 focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600 cursor-pointer";

  const abButton = () => (
    <span className="relative inline-flex items-center px-4 py-2 text-sm font-semibold text-gray-11 ring-1 ring-inset ring-gray-3 focus:outline-offset-0">
      ...
    </span>
  );
  const main = () => (
    <>
      <a
        onClick={() => handleChange(1)}
        aria-current="page"
        className={current === 1 ? activeClass : inactiveClass}
      >
        1
      </a>
      {current < 5 ? (
        <a
          onClick={() => handleChange(2)}
          className={current === 2 ? activeClass : inactiveClass}
        >
          2
        </a>
      ) : (
        abButton()
      )}
      {current > 4 ? (
        current < total - 3 ? (
          <a
            className={inactiveClass}
            onClick={() => handleChange(current - 1)}
          >
            {current - 1}
          </a>
        ) : (
          <a className={inactiveClass} onClick={() => handleChange(total - 4)}>
            {total - 4}
          </a>
        )
      ) : (
        <a
          onClick={() => handleChange(3)}
          className={current === 3 ? activeClass : inactiveClass}
        >
          {3}
        </a>
      )}
      {current > 4 ? (
        current < total - 2 ? (
          <a className={activeClass} onClick={() => handleChange(total - 2)}>
            {current}
          </a>
        ) : (
          <a className={inactiveClass} onClick={() => handleChange(total - 3)}>
            {total - 3}
          </a>
        )
      ) : (
        <a
          onClick={() => handleChange(4)}
          className={current === 4 ? activeClass : inactiveClass}
        >
          {4}
        </a>
      )}
      {current > total - 3 ? (
        <a
          className={current === total - 2 ? activeClass : inactiveClass}
          onClick={() => handleChange(total - 2)}
        >
          {total - 2}
        </a>
      ) : current > 3 ? (
        <a className={inactiveClass} onClick={() => handleChange(current + 1)}>
          {current + 1}
        </a>
      ) : (
        abButton()
      )}

      {current > total - 4 ? (
        <a
          className={current === total - 1 ? activeClass : inactiveClass}
          onClick={() => handleChange(total - 1)}
        >
          {total - 1}
        </a>
      ) : current < 4 ? (
        <a
          className={current === total - 1 ? activeClass : inactiveClass}
          onClick={() => handleChange(total - 1)}
        >
          {total - 1}
        </a>
      ) : (
        abButton()
      )}

      <a
        onClick={() => handleChange(total)}
        className={current === total ? activeClass : inactiveClass}
      >
        {total}
      </a>
    </>
  );

  const simple = () => (
    <>
      {new Array(total).fill(0).map((prop, key) => (
        <a
          key={key}
          onClick={() => handleChange(key + 1)}
          aria-current="page"
          className={current === key + 1 ? activeClass : inactiveClass}
        >
          {key + 1}
        </a>
      ))}
    </>
  );
  return (
    <div className="flex items-center justify-between border-t border-gray-2 bg-white px-4 py-3 sm:px-6">
      <div className="flex flex-1 justify-between sm:hidden">
        <a
          onClick={handlePrev}
          className="relative inline-flex cursor-pointer items-center rounded-md border border-gray-3 bg-white px-4 py-2 text-sm font-medium text-gray-11 hover:bg-gray-1"
        >
          Previous
        </a>
        <a
          onClick={handleNext}
          className="relative ml-3 inline-flex cursor-pointer items-center rounded-md border border-gray-3 bg-white px-4 py-2 text-sm font-medium text-gray-11 hover:bg-gray-1"
        >
          Next
        </a>
      </div>
      <div className="hidden sm:flex sm:flex-1 sm:items-center sm:justify-between">
        <div></div>
        <div>
          <nav
            className="isolate inline-flex -space-x-px rounded-md shadow-sm"
            aria-label="Pagination"
          >
            <a
              onClick={handlePrev}
              className="relative inline-flex cursor-pointer items-center rounded-l-md p-2 text-gray-4 ring-1 ring-inset ring-gray-3 hover:bg-gray-1 focus:z-20 focus:outline-offset-0"
            >
              <span className="sr-only">Previous</span>
              <ChevronLeftIcon className="h-5 w-5" aria-hidden="true" />
            </a>
            {
              //show all buttons when total is smaller than 7
              //show 7 buttons when total is bigger
              total > 6 ? main() : simple()
            }

            <a
              onClick={handleNext}
              className="relative inline-flex cursor-pointer items-center rounded-r-md p-2 text-gray-4 ring-1 ring-inset ring-gray-3 hover:bg-gray-1 focus:z-20 focus:outline-offset-0"
            >
              <span className="sr-only">Next</span>
              <ChevronRightIcon className="h-5 w-5" aria-hidden="true" />
            </a>
          </nav>
        </div>
      </div>
    </div>
  );
}
