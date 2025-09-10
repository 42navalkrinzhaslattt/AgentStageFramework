import React from "react";
import styled from "styled-components";

const ImpactContainer = styled.div`
  background: #1a1a1a;
  border-radius: 8px;
  padding: 1.5rem;
  margin: 1rem 0;
  border: 1px solid #333;
  border-left: 4px solid #ff9800;
`;

const ImpactTitle = styled.h3`
  color: #ff9800;
  margin: 0 0 1rem 0;
  font-size: 1.1rem;
  font-weight: 600;
`;

const MetricItem = styled.div`
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin: 0.8rem 0;
  padding: 0.8rem;
  background: #252525;
  border-radius: 6px;
  border-left: 3px solid ${(props) => props.color};
`;

const MetricName = styled.span`
  color: #cccccc;
  font-weight: 500;
`;

const MetricChange = styled.span`
  color: ${(props) => props.color};
  font-weight: 600;
  font-size: 1rem;
`;

const getChangeColor = (change) => {
  if (change > 0) return "#4CAF50";
  if (change < 0) return "#f44336";
  return "#9e9e9e";
};

const formatChange = (change) => {
  if (change > 0) return `+${change}`;
  return change.toString();
};

function ImpactDisplay({ impacts }) {
  if (!impacts || impacts.length === 0) return null;

  return (
    <ImpactContainer>
      <ImpactTitle>📈 Impact on Metrics</ImpactTitle>
      {impacts.map((impact, index) => {
        const color = getChangeColor(impact.change);
        return (
          <MetricItem key={index} color={color}>
            <MetricName>{impact.metric}</MetricName>
            <MetricChange color={color}>
              {formatChange(impact.change)}
            </MetricChange>
          </MetricItem>
        );
      })}
    </ImpactContainer>
  );
}

export default ImpactDisplay;
