package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"

	"google.golang.org/grpc"
	"open-match.dev/open-match/pkg/pb"
)

type AllocatePort struct {
	Name string `json:"name"`
	Port int    `json:"port"`
}
type AllocateStatus struct {
	State          string         `json:"state"`
	GameServerName string         `json:"gameServerName"`
	Ports          []AllocatePort `json:"ports"`
	Address        string         `json:"address"`
	NodeName       string         `json:"nodeName"`
}
type AllocateResponce struct {
	Status AllocateStatus `json:"status"`
}

// The Director in this tutorial continously polls Open Match for the Match
// Profiles and makes random assignments for the Tickets in the returned matches.

const (
	// The endpoint for the Open Match Backend service.
	omBackendEndpoint = "om-backend.open-match.svc.cluster.local:50505"
	// The Host and Port for the Match Function service endpoint.
	functionHostName       = "matchfunction.openmatch.svc.cluster.local"
	functionPort     int32 = 50502

	// The Host and Port for the AllocateService endpoint.
	allocateHostName = "http://fleet-allocator-endpoint.default.svc.cluster.local/address"
	allocateKey      = "v1GameClientKey"
	allocatePass     = "EAEC945C371B2EC361DE399C2F11E"
)

func main() {
	// Connect to Open Match Backend.
	conn, err := grpc.Dial(omBackendEndpoint, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to Open Match Backend, got %s", err.Error())
	}

	defer conn.Close()
	be := pb.NewBackendServiceClient(conn)

	// Generate the profiles to fetch matches for.
	profiles := generateProfiles()
	log.Printf("Fetching matches for %v profiles", len(profiles))

	for range time.Tick(time.Second * 1) {
		// Fetch matches for each profile and make random assignments for Tickets in
		// the matches returned.
		var wg sync.WaitGroup
		for _, p := range profiles {
			wg.Add(1)
			go func(wg *sync.WaitGroup, p *pb.MatchProfile) {
				defer wg.Done()
				matches, err := fetch(be, p)
				if err != nil {
					log.Printf("Failed to fetch matches for profile %v, got %s", p.GetName(), err.Error())
					return
				}

				if len(matches) > 0 {
					log.Printf("Generated %v matches for profile %v", len(matches), p.GetName())
				}
				if err := assign(be, matches); err != nil {
					log.Printf("Failed to assign servers to matches, got %s", err.Error())
					return
				}
			}(&wg, p)
		}

		wg.Wait()
	}
}

func fetch(be pb.BackendServiceClient, p *pb.MatchProfile) ([]*pb.Match, error) {
	req := &pb.FetchMatchesRequest{
		Config: &pb.FunctionConfig{
			Host: functionHostName,
			Port: functionPort,
			Type: pb.FunctionConfig_GRPC,
		},
		Profile: p,
	}

	stream, err := be.FetchMatches(context.Background(), req)
	if err != nil {
		log.Println()
		return nil, err
	}

	var result []*pb.Match
	for {
		resp, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			return nil, err
		}

		result = append(result, resp.GetMatch())
	}

	return result, nil
}

func assign(be pb.BackendServiceClient, matches []*pb.Match) error {
	for _, match := range matches {
		ticketIDs := []string{}
		for _, t := range match.GetTickets() {
			ticketIDs = append(ticketIDs, t.Id)
		}

		// Request Connection to AllocateService.
		aloReq, err := http.NewRequest("GET", allocateHostName, nil)
		if err != nil {
			return err
		}
		aloReq.SetBasicAuth(allocateKey, allocatePass)

		client := new(http.Client)
		resp, err := client.Do(aloReq)
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		byteArray, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		var alo AllocateResponce
		json.Unmarshal(byteArray, &alo)
		var conn string
		conn = fmt.Sprintf("%s:%d", alo.Status.Address, alo.Status.Ports[0].Port)

		req := &pb.AssignTicketsRequest{
			TicketIds: ticketIDs,
			Assignment: &pb.Assignment{
				Connection: conn,
			},
		}

		if _, err := be.AssignTickets(context.Background(), req); err != nil {
			return fmt.Errorf("AssignTickets failed for match %v, got %w", match.GetMatchId(), err)
		}

		log.Printf("Assigned server %v to match %v", conn, match.GetMatchId())
	}

	return nil
}
