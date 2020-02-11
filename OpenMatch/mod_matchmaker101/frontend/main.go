package main

// The Frontend in this tutorial continously creates Tickets in batches in Open Match.

import (
	"context"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo"

	"google.golang.org/grpc"
	"open-match.dev/open-match/pkg/pb"
)

type matchResponce struct {
	IP   string `json:"ip"`
	Port string `json:"port"`
}

const (
	// The endpoint for the Open Match Frontend service.
	omFrontendEndpoint = "om-frontend.open-match.svc.cluster.local:50504"
	// Number of tickets created per iteration
	ticketsPerIter = 20
)

var fe pb.FrontendServiceClient

func main() {
	// Connect to Open Match Frontend.
	conn, err := grpc.Dial(omFrontendEndpoint, grpc.WithInsecure())
	if err != nil {
		log.Fatalf("Failed to connect to Open Match, got %v", err)
	}

	defer conn.Close()
	fe = pb.NewFrontendServiceClient(conn)

	// create REST
	e := echo.New()
	e.GET("/frontend/:gamemode", handleGetMatch)
	e.Start(":80")
}

func handleGetMatch(c echo.Context) error {
	matchRes := new(matchResponce)
	if err := c.Bind(matchRes); err != nil {
		log.Fatalf("Failed to echo Bind, got %v", err)
		return c.JSON(http.StatusInternalServerError, matchRes)
	}

	// Create Ticket.
	gamemode := c.Param("gamemode")
	req := &pb.CreateTicketRequest{
		Ticket: makeTicket(gamemode),
	}
	resp, err := fe.CreateTicket(context.Background(), req)
	if err != nil {
		log.Fatalf("Failed to CreateTicket, got %v", err)
		return c.JSON(http.StatusInternalServerError, matchRes)
	}
	t := resp.Ticket
	log.Printf("Create Ticket: %v", t.GetId())

	// Polling TicketAssignment.
	for {
		got, err := fe.GetTicket(context.Background(), &pb.GetTicketRequest{TicketId: t.GetId()})
		if err != nil {
			log.Fatalf("Failed to GetTicket, got %v", err)
			return c.JSON(http.StatusInternalServerError, matchRes)
		}

		if got.GetAssignment() != nil {
			log.Printf("Ticket %v got assignment %v", got.GetId(), got.GetAssignment())
			conn := got.GetAssignment().Connection
			slice := strings.Split(conn, ":")
			matchRes.IP = slice[0]
			matchRes.Port = slice[1]
			break
		}
		time.Sleep(time.Second * 1)
	}

	_, err = fe.DeleteTicket(context.Background(), &pb.DeleteTicketRequest{TicketId: t.GetId()})
	if err != nil {
		log.Fatalf("Failed to Delete Ticket %v, got %s", t.GetId(), err.Error())
	}
	return c.JSON(http.StatusOK, matchRes)
}
