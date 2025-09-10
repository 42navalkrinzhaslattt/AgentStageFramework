import React from "react";
import styled from "styled-components";

const Panel = styled.div`
  background: rgba(255, 215, 0, 0.1);
  border: 2px solid #ffd700;
  border-radius: 15px;
  padding: 2rem;
  backdrop-filter: blur(10px);
`;

const Title = styled.h3`
  margin: 0 0 1rem 0;
  color: #ffd700;
  text-align: center;
`;

const SelectedChoice = styled.div`
  background: rgba(255, 255, 255, 0.1);
  border-radius: 10px;
  padding: 1rem;
  margin-bottom: 1.5rem;
  border-left: 4px solid #ffd700;
`;

const ChoiceLabel = styled.div`
  font-weight: bold;
  margin-bottom: 0.5rem;
  color: #ffd700;
`;

const ChoiceText = styled.div`
  font-size: 1rem;
  line-height: 1.4;
`;

const ReasoningSection = styled.div`
  margin-bottom: 1.5rem;
`;

const ReasoningLabel = styled.label`
  display: block;
  margin-bottom: 0.5rem;
  font-weight: bold;
  color: #ffd700;
`;

const ReasoningTextarea = styled.textarea`
  width: 100%;
  min-height: 100px;
  padding: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.3);
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.1);
  color: white;
  font-size: 1rem;
  font-family: inherit;
  resize: vertical;

  &::placeholder {
    color: rgba(255, 255, 255, 0.6);
  }

  &:focus {
    outline: none;
    border-color: #ffd700;
    background: rgba(255, 255, 255, 0.15);
  }
`;

const ButtonContainer = styled.div`
  display: flex;
  gap: 1rem;
  justify-content: center;
`;

const Button = styled.button`
  padding: 0.8rem 1.5rem;
  border: none;
  border-radius: 25px;
  font-size: 1rem;
  font-weight: bold;
  cursor: pointer;
  transition: all 0.3s ease;

  &:disabled {
    opacity: 0.6;
    cursor: not-allowed;
  }
`;

const ConfirmButton = styled(Button)`
  background: linear-gradient(45deg, #4caf50, #45a049);
  color: white;

  &:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 15px rgba(76, 175, 80, 0.3);
  }
`;

const CancelButton = styled(Button)`
  background: linear-gradient(45deg, #f44336, #da190b);
  color: white;

  &:hover:not(:disabled) {
    transform: translateY(-2px);
    box-shadow: 0 4px 15px rgba(244, 67, 54, 0.3);
  }
`;

const LoadingSpinner = styled.div`
  width: 20px;
  height: 20px;
  border: 2px solid #ffffff;
  border-top: 2px solid transparent;
  border-radius: 50%;
  animation: spin 1s linear infinite;
  margin-right: 10px;

  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }
`;

const ButtonContent = styled.div`
  display: flex;
  align-items: center;
  justify-content: center;
`;

function ChoicePanel({
  choice,
  reasoning,
  onReasoningChange,
  onConfirm,
  onCancel,
  loading,
}) {
  if (!choice) return null;

  return (
    <Panel>
      <Title>Decision Confirmation</Title>

      <SelectedChoice>
        <ChoiceLabel>Selected Option {choice.optionIndex + 1}:</ChoiceLabel>
        <ChoiceText>{choice.option}</ChoiceText>
      </SelectedChoice>

      <ReasoningSection>
        <ReasoningLabel htmlFor="reasoning">
          Decision Reasoning (optional):
        </ReasoningLabel>
        <ReasoningTextarea
          id="reasoning"
          value={reasoning}
          onChange={(e) => onReasoningChange(e.target.value)}
          placeholder="Explain why you chose this option. This will help in evaluating the consequences of your decision..."
          disabled={loading}
        />
      </ReasoningSection>

      <ButtonContainer>
        <CancelButton onClick={onCancel} disabled={loading}>
          Cancel
        </CancelButton>
        <ConfirmButton onClick={onConfirm} disabled={loading}>
          <ButtonContent>
            {loading && <LoadingSpinner />}
            {loading ? "Processing..." : "Confirm"}
          </ButtonContent>
        </ConfirmButton>
      </ButtonContainer>
    </Panel>
  );
}

export default ChoicePanel;
