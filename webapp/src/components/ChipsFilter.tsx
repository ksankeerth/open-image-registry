import { Checkbox, CheckboxChangeEvent } from "primereact/checkbox";
import { Chip } from "primereact/chip";
import { OverlayPanel } from "primereact/overlaypanel";
import React, { useRef, useState } from "react";

export type ChipFilterOption = {
  label: string;
  value: string;
};

export type ChipFilterProps = {
  filterOptions: ChipFilterOption[];
  handleFilterChange: (options: string[]) => void;
};

const ChipsFilter = (props: ChipFilterProps) => {
  const opRef = useRef<OverlayPanel>(null);
  const [showOptions, setShowOptions] = useState<boolean>(false);
  const [selectedOptions, setSelectedOptions] = useState<Set<string>>(
    new Set(props.filterOptions.map((op) => op.value))
  );

  const handleFilterChange = (e: CheckboxChangeEvent) => {
    if (e.checked) {
      setSelectedOptions((current) => {
        const newSet = new Set(current);
        newSet.add(e.value);
        props.handleFilterChange(Array.from(newSet))
        return newSet;
      });
    } else {
      setSelectedOptions((current) => {
        const newSet = new Set(current);
        newSet.delete(e.value);
        props.handleFilterChange(Array.from(newSet))
        return newSet;
      });
    }


  }

  return (
    <React.Fragment>
      {/*  border-1 border-solid border-round-lg border-teal-100 */}
      <div
        className="p-2 m-0
        
        flex justify-content-between align-items-center
        hover:bg-white hover:border-1"
        onClick={(e) => {
          opRef.current?.toggle(e);
        }}
      >
        {!showOptions && (
          <span
            className="pi pi-chevron-down  text-xs pr-2 pl-2"
            onClick={() => setShowOptions((c) => !c)}
          />
        )}
        {showOptions && (
          <span
            className="pi pi-chevron-up  text-xs pr-2 pl-2"
            onClick={() => setShowOptions((c) => !c)}
          />
        )}

        <div className="flex ">
          {selectedOptions.size != 0 &&
            Array.from(selectedOptions).map((value: string) => {
              return (
                <Chip
                  className="m-1 text-xs"
                  label={
                    props.filterOptions.find((op) => op.value == value)?.label
                  }
                />
              );
            })}
        </div>

        <OverlayPanel ref={opRef} className="p-0 m-0">
          <div className="flex flex-column text-xs">
            {props.filterOptions.map((node: ChipFilterOption) => (
              <div className="flex flex-row gap-5 p-1 align-items-center">
                <Checkbox
                  value={node.value}
                  checked={selectedOptions.has(node.value)}
                  onChange={handleFilterChange}
                />
                <div>{node.label}</div>
              </div>
            ))}
          </div>
        </OverlayPanel>
      </div>
    </React.Fragment>
  );
};

export default ChipsFilter;