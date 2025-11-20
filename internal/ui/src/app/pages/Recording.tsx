import React from "react";
import { useParams, useNavigate } from "react-router";
import RecordingDetail from "../components/RecordingDetail";

export default function Recording() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  if (!id) {
    navigate("/recordings");
    return null;
  }

  return (
    <RecordingDetail recordingId={id} />
  );
}
