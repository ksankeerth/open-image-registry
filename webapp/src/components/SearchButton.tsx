import { InputText } from 'primereact/inputtext';
import React from 'react';

type SearchButtonProps = {
  placeholder: string;
  handleSearch: (searchTerm: string) => void;
};

const SearchButton = (props: SearchButtonProps) => {
  const [isFocused, setIsFocused] = React.useState(false);

  return (
    <div
      className={`
        flex align-items-center gap-2
        bg-white border-round-3xl
        px-3 py-2 w-20rem
        transition-all transition-duration-200
        ${isFocused ? 'shadow-3' : 'shadow-1'}
      `}
      style={{
        boxShadow: isFocused
          ? '0 0 0 3px rgba(109, 213, 176, 0.1), 0 2px 8px rgba(0, 0, 0, 0.1)'
          : '0 1px 3px rgba(0, 0, 0, 0.08)',
      }}
    >
      <span
        className={`pi pi-search text-sm transition-colors transition-duration-200 ${isFocused ? 'text-teal-400' : 'text-gray-500'}`}
      />

      <InputText
        type="text"
        className="border-none text-sm p-0 flex-grow-1"
        placeholder={props.placeholder}
        onChange={(e) => props.handleSearch(e.target.value)}
        onFocus={() => setIsFocused(true)}
        onBlur={() => setIsFocused(false)}
        style={{
          outline: 'none',
          boxShadow: 'none',
        }}
      />
    </div>
  );
};

export default SearchButton;