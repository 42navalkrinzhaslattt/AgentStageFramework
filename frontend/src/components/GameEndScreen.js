import React from "react";
import styled from "styled-components";

const EndScreenContainer = styled.div`
  background: #1a1a1a;
  border-radius: 12px;
  padding: 2rem;
  margin: 2rem 0;
  border: 2px solid #4caf50;
  text-align: center;
  box-shadow: 0 4px 20px rgba(76, 175, 80, 0.2);
`;

const GameOverTitle = styled.h1`
  color: #4caf50;
  font-size: 2.5rem;
  margin: 0 0 1rem 0;
  font-weight: 700;
`;

const FinalScoreContainer = styled.div`
  background: #252525;
  border-radius: 8px;
  padding: 1.5rem;
  margin: 1.5rem 0;
  border: 1px solid #333;
`;

const ScoreTitle = styled.h2`
  color: #ff9800;
  margin: 0 0 1rem 0;
  font-size: 1.5rem;
`;

const MetricRow = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 0.8rem 0;
  padding: 0.8rem;
  background: #1a1a1a;
  border-radius: 6px;
  border-left: 3px solid #4a9eff;
`;

const MetricLabel = styled.span`
  color: #cccccc;
  font-weight: 500;
  font-size: 1rem;
`;

const MetricValue = styled.span`
  color: #4caf50;
  font-weight: 600;
  font-size: 1.1rem;
`;

const SummaryText = styled.p`
  color: #cccccc;
  font-size: 1.1rem;
  line-height: 1.6;
  margin: 1.5rem 0;
`;

const RestartButton = styled.button`
  background: #4caf50;
  color: white;
  border: none;
  padding: 1rem 2rem;
  border-radius: 8px;
  font-size: 1.1rem;
  font-weight: 600;
  cursor: pointer;
  transition: background-color 0.3s;

  &:hover {
    background: #45a049;
  }
`;

function GameEndScreen({ finalMetrics, summary, onRestart }) {
  return (
    <EndScreenContainer>
      <GameOverTitle>🎮 Game Complete!</GameOverTitle>

      <FinalScoreContainer>
        <ScoreTitle>📊 Final Metrics</ScoreTitle>
        {finalMetrics &&
          finalMetrics.map((metric, index) => (
            <MetricRow key={index}>
              <MetricLabel>{metric.name}</MetricLabel>
              <MetricValue>{metric.value}</MetricValue>
            </MetricRow>
          ))}
      </FinalScoreContainer>

      {summary && <SummaryText>{summary}</SummaryText>}

      {onRestart && (
        <RestartButton onClick={onRestart}>🔄 Play Again</RestartButton>
      )}
    </EndScreenContainer>
  );
}

export default GameEndScreen;
