import { useEffect, useCallback, useState, useRef } from 'react';
import { useClickAway } from 'ahooks';
import './index.less';

const SelectLange = ({ value = '02', options = [], onChange }) => {
  const [isOpen, setIsOpen] = useState(false);
  const dropdownRef = useRef(null);
  const [selectedOption, setSelectedOption] = useState(options.find(opt => opt.value === value) || options[0]);

  useClickAway(() => {
    setIsOpen(false);
  }, dropdownRef);

  const handleToggle = useCallback(() => {
    setIsOpen(prev => !prev);
  }, []);

  const handleSelect = useCallback((option) => {
    setSelectedOption(option);
    setIsOpen(false);
    onChange?.(option);
  }, [onChange]);

  useEffect(() => {
    const option = options.find(opt => opt.value === value);
    if (option) {
      setSelectedOption(option);
    }
  }, [value, options]);

  return (
    <div className="select-lange" ref={dropdownRef}>
      <div className="select-trigger" onClick={handleToggle}>
        <span className="selected-value">{selectedOption?.label}</span>
        <span className={`selected-icon ${!isOpen ? 'selected-icon-up' : 'selected-icon-down'}`} />
      </div>
      {isOpen && (
        <div className="options-container">
          {options.map((option, index) => (
            <div
              key={option.value}
              className={`option ${option.value === selectedOption?.value ? 'active' : ''}`}
              onClick={() => handleSelect(option, index)}
            >
              {option.label}
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default SelectLange;
