package main

import (
	"open-match.dev/open-match/pkg/pb"
)

// generateProfiles generates test profiles for the matchmaker101 tutorial.
func generateProfiles() []*pb.MatchProfile {
	var profiles []*pb.MatchProfile
	modes := []string{"mode.demo", "mode.ctf", "mode.battleroyale"}
	for _, mode := range modes {
		profiles = append(profiles, &pb.MatchProfile{
			Name: "mode_based_profile",
			Pools: []*pb.Pool{
				{
					Name: "pool_mode_" + mode,
					TagPresentFilters: []*pb.TagPresentFilter{
						{
							Tag: mode,
						},
					},
				},
			},
		},
		)
	}

	return profiles
}
