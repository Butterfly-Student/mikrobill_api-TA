package ppp

import (
	"context"
	"fmt"

	"github.com/palantir/stacktrace"

	"MikrOps/internal/model"
)

// TODO: Make this configurable via environment variable or settings
const PPPOnUpScript = `:local apiUrl "http://192.168.100.2:8080/v1/callbacks/pppoe/up"

:local user $user
:local callerId $"caller-id"
:local interfaceName $interface
:local localAddr $"local-address"
:local remoteAddr $"remote-address"
:local service $service

:local jsonData ("{\"name\":\"" . $user . \
               "\",\"caller_id\":\"" . $callerId . \
               "\",\"interface\":\"" . $interfaceName . \
               "\",\"local_address\":\"" . $localAddr . \
               "\",\"remote_address\":\"" . $remoteAddr . \
               "\",\"service\":\"" . $service . "\"}")

/tool fetch url=$apiUrl \
    http-method=post \
    http-header-field="Content-Type: application/json" \
    http-data=$jsonData \
    keep-result=no`

const PPPOnDownScript = `:local apiUrl "http://192.168.100.2:8080/v1/callbacks/pppoe/down"

:local user $user
:local callerId $"caller-id"
:local interfaceName $interface
:local localAddr $"local-address"
:local remoteAddr $"remote-address"
:local service $service

:local jsonData ("{\"name\":\"" . $user . \
               "\",\"caller_id\":\"" . $callerId . \
               "\",\"interface\":\"" . $interfaceName . \
               "\",\"local_address\":\"" . $localAddr . \
               "\",\"remote_address\":\"" . $remoteAddr . \
               "\",\"service\":\"" . $service . "\"}")

/tool fetch url=$apiUrl \
    http-method=post \
    http-header-field="Content-Type: application/json" \
    http-data=$jsonData \
    keep-result=no`

// --- Profiles ---

func (d *PPPDomain) MikrotikCreateProfile(ctx context.Context, input model.PPPProfileInput) (*model.PPPProfile, error) {
	client, err := d.getActiveClient(ctx)

	if err != nil {
		return nil, err
	}
	defer client.Close()

	_, err = client.RunArgs("/ppp/profile/add", map[string]string{
		"name":           input.Name,
		"local-address":  input.LocalAddress,
		"remote-address": input.RemoteAddress,
		"rate-limit":     input.RateLimit,
		"only-one":       input.OnlyOne,
		"comment":        input.Comment,
		"on-up":          PPPOnUpScript,
		"on-down":        PPPOnDownScript,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to create ppp profile")
	}

	return &model.PPPProfile{Name: input.Name}, nil
}

func (d *PPPDomain) MikrotikGetProfile(ctx context.Context, id string) (*model.PPPProfile, error) {
	client, err := d.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	reply, err := client.RunArgs("/ppp/profile/print", map[string]string{
		"?.id": id,
	})
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to get ppp profile")
	}
	if len(reply.Re) == 0 {
		return nil, fmt.Errorf("ppp profile not found")
	}

	r := reply.Re[0].Map
	return &model.PPPProfile{
		Name:      r["name"],
		RateLimit: r["rate-limit"],
	}, nil
}

func (d *PPPDomain) MikrotikUpdateProfile(ctx context.Context, id string, input model.PPPProfileInput) (*model.PPPProfile, error) {
	client, err := d.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	args := map[string]string{".id": id}
	args["name"] = input.Name
	args["on-up"] = PPPOnUpScript
	args["on-down"] = PPPOnDownScript
	// ... other fields

	_, err = client.RunArgs("/ppp/profile/set", args)
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to update ppp profile")
	}

	return nil, nil
}

func (d *PPPDomain) MikrotikDeleteProfile(ctx context.Context, id string) error {
	client, err := d.getActiveClient(ctx)
	if err != nil {
		return err
	}
	defer client.Close()

	_, err = client.RunArgs("/ppp/profile/remove", map[string]string{
		".id": id,
	})
	if err != nil {
		return stacktrace.Propagate(err, "failed to delete ppp profile")
	}
	return nil
}

func (d *PPPDomain) MikrotikListProfiles(ctx context.Context) ([]model.PPPProfile, error) {
	client, err := d.getActiveClient(ctx)
	if err != nil {
		return nil, err
	}
	defer client.Close()

	reply, err := client.Run("/ppp/profile/print")
	if err != nil {
		return nil, stacktrace.Propagate(err, "failed to list ppp profiles")
	}

	var profiles []model.PPPProfile
	for _, re := range reply.Re {
		m := re.Map
		profiles = append(profiles, model.PPPProfile{
			Name:      m["name"],
			RateLimit: m["rate-limit"],
		})
	}
	return profiles, nil
}
