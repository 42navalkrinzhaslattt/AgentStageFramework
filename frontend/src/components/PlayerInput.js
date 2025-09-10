import React, { useState } from "react";
import styled from "styled-components";

const InputContainer = styled.div`
  background: #1a1a1a;
  border-radius: 8px;
  padding: 1.5rem;
  margin: 1rem 0;
  border: 1px solid #333;
`;

const InputLabel = styled.div`
  color: #ffffff;
  font-size: 0.95rem;
  margin-bottom: 1rem;
  font-weight: 500;
`;

const TextArea = styled.textarea`
  width: 100%;
  min-height: 100px;
  background: #252525;
  border: 1px solid #444;
  border-radius: 6px;
  padding: 1rem;
  color: #ffffff;
  font-size: 0.95rem;
  font-family: inherit;
  resize: vertical;

  &:focus {
    outline: none;
    border-color: #4a9eff;
    box-shadow: 0 0 0 2px rgba(74, 158, 255, 0.2);
  }

  &::placeholder {
    color: #888;
  }
`;

const ButtonContainer = styled.div`
  display: flex;
  justify-content: flex-end;
  margin-top: 1rem;
  gap: 0.5rem;
`;

const SubmitButton = styled.button`
  background: #4a9eff;
  color: white;
  border: none;
  padding: 0.8rem 1.5rem;
  border-radius: 6px;
  font-size: 0.9rem;
  font-weight: 500;
  cursor: pointer;
  transition: background-color 0.2s;

  &:hover:not(:disabled) {
    background: #3a8eef;
  }

  &:disabled {
    background: #666;
    cursor: not-allowed;
  }
`;

const ClearButton = styled.button`
  background: transparent;
  color: #888;
  border: 1px solid #444;
  padding: 0.8rem 1.5rem;
  border-radius: 6px;
  font-size: 0.9rem;
  cursor: pointer;
  transition: all 0.2s;

  &:hover {
    color: #ffffff;
    border-color: #666;
  }
`;

function PlayerInput({
  onSubmit,
  disabled = false,
  placeholder = "Enter your strategic response...",
}) {
  const [input, setInput] = useState("");

  const handleSubmit = () => {
    if (input.trim() && onSubmit) {
      onSubmit(input.trim());
      setInput("");
    }
  };

  const handleClear = () => {
    setInput("");
  };

  const handleKeyPress = (e) => {
    if (e.key === "Enter" && e.ctrlKey) {
      handleSubmit();
    }
  };

  return (
    <InputContainer>
      <InputLabel>Your strategic response (free-form):</InputLabel>

      <TextArea
        value={input}
        onChange={(e) => setInput(e.target.value)}
        onKeyPress={handleKeyPress}
        placeholder={placeholder}
        disabled={disabled}
      />

      <ButtonContainer>
        <ClearButton onClick={handleClear} disabled={disabled || !input.trim()}>
          Clear
        </ClearButton>
        <SubmitButton
          onClick={handleSubmit}
          disabled={disabled || !input.trim()}
        >
          Submit Response
        </SubmitButton>
      </ButtonContainer>
    </InputContainer>
  );
}

export default PlayerInput;
