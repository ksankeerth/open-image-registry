import { Checkbox, CheckboxChangeEvent } from 'primereact/checkbox';
import { Chip } from 'primereact/chip';
import { OverlayPanel } from 'primereact/overlaypanel';
import React, { useRef, useState } from 'react';

export type ChipFilterOption = {
  label: string;
  value: string;
  unselected_default?: boolean;
};

export type ChipFilterProps = {
  maxChipsPerRow?: number;
  filterOptions: ChipFilterOption[];
  handleFilterChange: (options: string[]) => void;
};

const ChipsFilter = (props: ChipFilterProps) => {
  const opRef = useRef<OverlayPanel>(null);
  const [selectedOptions, setSelectedOptions] = useState<Set<string>>(() => {
    return new Set(
      props.filterOptions
        .filter((option) => !option.unselected_default)
        .map((option) => option.value)
    );
  });

  const handleCheckboxChange = (e: CheckboxChangeEvent) => {
    setSelectedOptions((prevSelected) => {
      const newSet = new Set(prevSelected);

      if (e.checked) {
        newSet.add(e.value);
      } else {
        newSet.delete(e.value);
      }

      props.handleFilterChange(Array.from(newSet));

      return newSet;
    });
  };

  const renderChips = () => {
    if (selectedOptions.size === 0) {
      return null;
    }

    const selectedArray = Array.from(selectedOptions);

    // If maxChipsPerRow is specified, render in rows
    if (props.maxChipsPerRow) {
      const rows = Math.ceil(selectedArray.length / props.maxChipsPerRow);

      return (
        <div className="flex flex-column">
          {Array.from({ length: rows }).map((_, rowIndex) => {
            const start = rowIndex * props.maxChipsPerRow!;
            const end = start + props.maxChipsPerRow!;
            const rowItems = selectedArray.slice(start, end);

            return (
              <div className="flex flex-row" key={rowIndex}>
                {rowItems.map((value) => {
                  const option = props.filterOptions.find((op) => op.value === value);
                  return (
                    <Chip
                      key={value}
                      label={option?.label || value}
                      className="m-1 text-xs font-inter"
                      pt={{
                        root: {
                          className: 'bg-white',
                        },
                      }}
                    />
                  );
                })}
              </div>
            );
          })}
        </div>
      );
    }

    // Single row of chips
    return (
      <div className="flex flex-wrap">
        {selectedArray.map((value) => {
          const option = props.filterOptions.find((op) => op.value === value);
          return <Chip key={value} label={option?.label || value} className="m-1 text-xs" />;
        })}
      </div>
    );
  };

  return (
    <div
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === 'Enter' || e.key === ' ') {
          opRef.current?.toggle(e);
        }
      }}
      className="p-2 m-0 flex align-items-center gap-2 cursor-pointer hover:border-1 border-round"
      onClick={(e) => opRef.current?.toggle(e)}
    >
      {/* Toggle Icon */}
      <i
        // eslint-disable-next-line react-hooks/refs
        className={`pi ${opRef.current?.state?.overlayVisible ? 'pi-chevron-up' : 'pi-chevron-down'} text-xs`}
      />

      {/* Chips Display */}
      {renderChips()}

      {/* Filter Options Overlay */}
      <OverlayPanel ref={opRef} pt={{ content: { className: 'p-3' } }}>
        <div className="flex flex-column gap-2">
          {props.filterOptions.map((option) => (
            <div key={option.value} className="flex align-items-center gap-3">
              <Checkbox
                inputId={`filter-${option.value}`}
                value={option.value}
                checked={selectedOptions.has(option.value)}
                onChange={handleCheckboxChange}
              />
              <label htmlFor={`filter-${option.value}`} className="cursor-pointer text-sm">
                {option.label}
              </label>
            </div>
          ))}
        </div>
      </OverlayPanel>
    </div>
  );
};

export default ChipsFilter;
