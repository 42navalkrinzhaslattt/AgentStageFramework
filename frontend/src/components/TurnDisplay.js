import React from "react";
import styled from "styled-components";

const TurnContainer = styled.div`
  background: #1a1a1a;
  border-radius: 8px;
  padding: 1rem;
  margin: 1rem 0;
  border: 1px solid #333;
  text-align: center;
`;

const Separator = styled.div`
  color: #666;
  font-family: "Courier New", monospace;
  font-size: 0.9rem;
  margin: 0.5rem 0;
`;

const TurnTitle = styled.h2`
  color: #ffffff;
  font-size: 1.2rem;
  margin: 0.5rem 0;
  font-weight: 600;
`;

function TurnDisplay({ currentTurn, maxTurns }) {
  const separatorLine = "=".repeat(60);

  return (
    <TurnContainer>
      <Separator>{separatorLine}</Separator>
      <TurnTitle>
        🏛️ TURN {currentTurn} of {maxTurns}
      </TurnTitle>
      <Separator>{separatorLine}</Separator>
    </TurnContainer>
  );
}

export default TurnDisplay;
