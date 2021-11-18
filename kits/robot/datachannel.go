package robot

import (
	"encoding/json"
	"errors"

	log "github.com/doublemo/baa/cores/log/level"
	coresproto "github.com/doublemo/baa/cores/proto"
	"github.com/doublemo/baa/internal/proto/command"
	"github.com/doublemo/baa/internal/proto/kit"
	"github.com/doublemo/baa/internal/proto/pb"
	"github.com/doublemo/baa/kits/robot/session"
	grpcproto "github.com/golang/protobuf/proto"
	"github.com/pion/webrtc/v3"
)

func openDataChannel(peer session.Peer, c RobotConfig) error {
	dc, err := peer.CreateDataChannel(c.Datachannel)
	if err != nil {
		return err
	}

	dc.OnICEConnectionStateChange(iceConnectionStateChange(peer))
	dc.OnICECandidate(icecandidate(peer))

	datachannel, ok := dc.DataChannel()
	if !ok {
		return errors.New("Default channel creation failed")
	}

	datachannel.OnOpen(func() {
		// ok readyed
		log.Debug(Logger()).Log("action", "Readyed", "Robot", peer.ID())
		task, ok := peer.Params("Task")
		if !ok {
			log.Warn(Logger()).Log("action", "task", "error", "No task can be executed, the machine stops automatically", "Robot", peer.ID())
			peer.Close()
			return
		}

		tk, ok := task.(*pb.Robot_Start_Robot)
		if !ok {
			log.Warn(Logger()).Log("action", "task.(*pb.Robot_Start_Robot)", "error", "No task can be executed, the machine stops automatically", "Robot", peer.ID())
			peer.Close()
			return
		}

		if err := execTask(peer, tk, c); err != nil {
			log.Error(Logger()).Log("action", "execTask", "error", err)
			peer.Close()
		}
	})

	return doRequestDataChannelOffer(peer, false)
}

func doRequestDataChannelOffer(peer session.Peer, restartIce bool) error {
	dc := peer.DataChannel()
	offer, err := dc.Offer(restartIce)
	if err != nil {
		return err
	}

	description, err := json.Marshal(offer)
	if err != nil {
		return err
	}

	frame := &pb.Agent_Webrtc_Signal{
		Payload: &pb.Agent_Webrtc_Signal_Description{
			Description: description,
		},
	}

	bytes, err := grpcproto.Marshal(frame)
	if err != nil {
		return err
	}

	req := &coresproto.RequestBytes{
		Ver:     1,
		Cmd:     kit.Agent,
		SubCmd:  command.AgentDatachannel,
		Content: bytes,
		SeqID:   1,
	}

	r, err := req.Marshal()
	if err != nil {
		return err
	}

	return peer.Send(session.PeerMessagePayload{Data: r})
}

func doRequestDataChannelTrickle(peer session.Peer, trickle *webrtc.ICECandidate) error {
	description, err := json.Marshal(trickle.ToJSON())
	if err != nil {
		return err
	}

	frame := &pb.Agent_Webrtc_Signal{
		Payload: &pb.Agent_Webrtc_Signal_Trickle{
			Trickle: &pb.Agent_Webrtc_Trickle{
				Candidate: string(description),
			},
		},
	}

	bytes, err := grpcproto.Marshal(frame)
	if err != nil {
		return err
	}

	req := &coresproto.RequestBytes{
		Ver:     1,
		Cmd:     kit.Agent,
		SubCmd:  command.AgentDatachannel,
		Content: bytes,
		SeqID:   1,
	}

	r, err := req.Marshal()
	if err != nil {
		return err
	}
	return peer.Send(session.PeerMessagePayload{Data: r})
}

func datachannel(peer session.Peer, w coresproto.Response) error {
	if w.StatusCode() != 0 {
		return errors.New(string(w.Body()))
	}

	var frame pb.Agent_Webrtc_Signal
	{
		if err := grpcproto.Unmarshal(w.Body(), &frame); err != nil {
			return err
		}
	}

	switch payload := frame.Payload.(type) {
	case *pb.Agent_Webrtc_Signal_Description:
		return dataChannelDescription(peer, payload)

	case *pb.Agent_Webrtc_Signal_IceConnectionState:
		log.Info(Logger()).Log("IceConnectionState", payload.IceConnectionState)

	case *pb.Agent_Webrtc_Signal_Trickle:
		return dataChannelTrickle(peer, payload)
	}

	return nil
}

func dataChannelDescription(peer session.Peer, description *pb.Agent_Webrtc_Signal_Description) error {
	var desc webrtc.SessionDescription
	{
		if err := json.Unmarshal(description.Description, &desc); err != nil {
			return err
		}
	}

	if desc.Type == webrtc.SDPTypeAnswer {
		dc := peer.DataChannel()
		return dc.SetRemoteDescription(desc)
	}

	return nil
}

func dataChannelTrickle(peer session.Peer, trickle *pb.Agent_Webrtc_Signal_Trickle) error {
	dc := peer.DataChannel()
	var candidate webrtc.ICECandidateInit
	{
		if err := json.Unmarshal([]byte(trickle.Trickle.Candidate), &candidate); err != nil {
			return err
		}
	}
	return dc.AddICECandidate(candidate)
}

func iceConnectionStateChange(peer session.Peer) func(connectionState webrtc.ICEConnectionState) {
	return func(connectionState webrtc.ICEConnectionState) {
		switch connectionState {
		case webrtc.ICEConnectionStateDisconnected:
			if err := doRequestDataChannelOffer(peer, true); err != nil {
				log.Error(Logger()).Log("action", "ICEConnectionStateDisconnected", "error", err)
			}

		case webrtc.ICEConnectionStateClosed, webrtc.ICEConnectionStateFailed:
			log.Warn(Logger()).Log("action", "ICEConnectionStateClosed", "state", connectionState)
			peer.Close()

		case webrtc.ICEConnectionStateCompleted:
			log.Info(Logger()).Log("action", "ICEConnectionStateCompleted", "state", connectionState)

		case webrtc.ICEConnectionStateConnected:
			log.Info(Logger()).Log("action", "ICEConnectionStateConnected", "state", connectionState)
		}
	}
}

func icecandidate(peer session.Peer) func(*webrtc.ICECandidate) {
	return func(c *webrtc.ICECandidate) {
		if c == nil {
			return
		}

		if err := doRequestDataChannelTrickle(peer, c); err != nil {
			log.Error(Logger()).Log("action", "doRequestDataChannelTrickle", "error", err)
		}
	}
}
