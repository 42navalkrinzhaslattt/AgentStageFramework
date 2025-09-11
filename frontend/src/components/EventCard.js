import React from "react";
import styled from "styled-components";

const Card = styled.div`
  background: rgba(255, 255, 255, 0.1);
  border-radius: 15px;
  padding: 2rem;
  backdrop-filter: blur(10px);
  border: 1px solid rgba(255, 255, 255, 0.2);
  margin-bottom: 1rem;
`;

const EventImage = styled.img`
  width: 100%;
  height: auto;
  border-radius: 10px;
  margin-bottom: 1rem;
  border: 1px solid rgba(255, 255, 255, 0.2);
  display: block;
`;

const EventTitle = styled.h2`
  margin: 0 0 1rem 0;
  color: #ffd700;
  font-size: 1.5rem;
`;

const EventDescription = styled.p`
  margin-bottom: 2rem;
  line-height: 1.6;
  font-size: 1.1rem;
  white-space: pre-wrap;
`;

const CategoryBadge = styled.span`
  display: inline-block;
  padding: 0.3rem 0.8rem;
  background: ${(props) => getCategoryColor(props.category)};
  color: white;
  border-radius: 20px;
  font-size: 0.8rem;
  font-weight: bold;
  margin-bottom: 1rem;
`;

const SeverityIndicator = styled.div`
  display: flex;
  align-items: center;
  margin-bottom: 1rem;
  font-size: 0.9rem;
`;

const SeverityDots = styled.div`
  display: flex;
  gap: 3px;
  margin-left: 0.5rem;
`;

const SeverityDot = styled.div`
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: ${(props) =>
    props.active ? getSeverityColor(props.severity) : "rgba(255,255,255,0.3)"};
`;

const OptionsContainer = styled.div`
  display: flex;
  flex-direction: column;
  gap: 0.8rem;
`;

const OptionButton = styled.button`
  background: ${(props) =>
    props.selected ? "rgba(255, 215, 0, 0.2)" : "rgba(255,255,255,0.1)"};
  border: 2px solid
    ${(props) => (props.selected ? "#ffd700" : "rgba(255,255,255,0.2)")};
  color: white;
  padding: 1rem;
  border-radius: 10px;
  cursor: ${(props) => (props.disabled ? "not-allowed" : "pointer")};
  transition: all 0.3s ease;
  text-align: left;
  font-size: 1rem;
  opacity: ${(props) => (props.disabled ? 0.6 : 1)};

  &:hover:not(:disabled) {
    background: rgba(255, 215, 0, 0.1);
    border-color: #ffd700;
    transform: translateY(-1px);
  }
`;

const OptionsTitle = styled.h3`
  margin: 1.5rem 0 1rem 0;
  color: #ffd700;
`;

function getCategoryColor(category) {
  const colors = {
    economy: "#4CAF50",
    security: "#F44336",
    diplomacy: "#2196F3",
    environment: "#8BC34A",
    domestic: "#FF9800",
    military: "#9C27B0",
    social: "#00BCD4",
    tech: "#607D8B",
  };
  return colors[category] || "#757575";
}

function getSeverityColor(severity) {
  if (severity <= 3) return "#4CAF50";
  if (severity <= 6) return "#FF9800";
  return "#F44336";
}

function getCategoryName(category) {
  const names = {
    economy: "Economy",
    security: "Security",
    diplomacy: "Diplomacy",
    environment: "Environment",
    domestic: "Domestic Policy",
    military: "Military Affairs",
    social: "Social Sphere",
    tech: "Technology",
  };
  return names[category] || category;
}

function EventCard({ event, onChoiceSelect, selectedChoice, disabled }) {
  if (!event) return null;

  return (
    <Card>
      <CategoryBadge category={event.category}>
        {getCategoryName(event.category)}
      </CategoryBadge>

      <SeverityIndicator>
        Crisis Level:
        <SeverityDots>
          {[...Array(10)].map((_, i) => (
            <SeverityDot
              key={i}
              active={i < event.severity}
              severity={event.severity}
            />
          ))}
        </SeverityDots>
        <span style={{ marginLeft: "0.5rem", fontWeight: "bold" }}>
          {event.severity}/10
        </span>
      </SeverityIndicator>

      {event.imageUrl && (
        <EventImage src={event.imageUrl} alt={event.title || "Event image"} />
      )}

      <EventTitle>{event.title}</EventTitle>
      <EventDescription>{event.description}</EventDescription>

      <OptionsTitle>Action Options:</OptionsTitle>
      <OptionsContainer>
        {event.options &&
          event.options.map((option, index) => (
            <OptionButton
              key={index}
              selected={selectedChoice && selectedChoice.optionIndex === index}
              disabled={disabled}
              onClick={() => !disabled && onChoiceSelect(index, option)}
            >
              <strong>{index + 1}.</strong> {option}
            </OptionButton>
          ))}
      </OptionsContainer>
    </Card>
  );
}

export default EventCard;
