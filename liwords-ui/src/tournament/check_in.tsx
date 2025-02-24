import { Button, Col, Divider, message, Row } from 'antd';
import axios from 'axios';
import React, { useMemo } from 'react';
import { toAPIUrl } from '../api/api';
import {
  useLoginStateStoreContext,
  useTournamentStoreContext,
} from '../store/store';

// I did not find a design for this, but it is trial functionality in order
// to keep the tournament running smoothly.

export const CheckIn = () => {
  const { tournamentContext } = useTournamentStoreContext();

  const { loginState } = useLoginStateStoreContext();

  // Only registered players can check in.
  const checkedIn = useMemo(() => {
    if (!tournamentContext.competitorState.division) {
      return false;
    }
    const division =
      tournamentContext.divisions[tournamentContext.competitorState.division];
    // return division.checkedInPlayers.has(
    //   loginState.userID + ':' + loginState.username
    // );
    // XXX: TEMP CODE so this thing compiles -- FIX ME!!
    return division !== null;
  }, [
    loginState.username,
    loginState.userID,
    tournamentContext.competitorState.division,
    tournamentContext.divisions,
  ]);

  if (!tournamentContext.competitorState.isRegistered) {
    return null;
  }
  if (checkedIn) {
    return null;
  }

  const checkin = () => {
    axios
      .post<{}>(toAPIUrl('tournament_service.TournamentService', 'CheckIn'), {
        id: tournamentContext.metadata?.getId(),
      })
      .then((resp) => {
        message.info({
          content: 'You are checked in.',
          duration: 3,
        });
      })
      .catch((err) => {
        message.error({
          content: 'Error checking in: ' + err.response?.data?.msg,
          duration: 5,
        });
      });
  };

  return (
    <Row>
      <Col offset={10}>
        <Button
          onClick={checkin}
          type="primary"
          size="large"
          style={{ marginTop: 54, marginBottom: 58 }}
        >
          CHECK IN
        </Button>
      </Col>
      <Divider />
    </Row>
  );
};
