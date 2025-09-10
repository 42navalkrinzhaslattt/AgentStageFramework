import React from "react";
import styled from "styled-components";

const ResultContainer = styled.div`
  background: #1a1a1a;
  border-radius: 8px;
  padding: 1.5rem;
  margin: 1rem 0;
  border: 1px solid #333;
  border-left: 4px solid #4caf50;
`;

const ResultItem = styled.div`
  margin: 1rem 0;
  color: #ffffff;
  font-size: 0.95rem;
  line-height: 1.5;
`;

const ResultLabel = styled.span`
  color: #4a9eff;
  font-weight: 600;
  margin-right: 0.5rem;
`;

const ChoiceText = styled.div`
  background: #252525;
  padding: 1rem;
  border-radius: 6px;
  margin: 0.5rem 0;
  border-left: 3px solid #4caf50;
  color: #cccccc;
`;

const EvaluationText = styled.div`
  background: #252525;
  padding: 1rem;
  border-radius: 6px;
  margin: 0.5rem 0;
  border-left: 3px solid #4a9eff;
  color: #cccccc;
`;

function ResultDisplay({ choice, evaluation }) {
  if (!choice && !evaluation) return null;

  return (
    <ResultContainer>
      {choice && (
        <ResultItem>
          <ResultLabel>🎯 You chose:</ResultLabel>
          <ChoiceText>{choice}</ChoiceText>
        </ResultItem>
      )}

      {evaluation && (
        <ResultItem>
          <ResultLabel>📊 Evaluation:</ResultLabel>
          <EvaluationText>{evaluation}</EvaluationText>
        </ResultItem>
      )}
    </ResultContainer>
  );
}

export default ResultDisplay;
