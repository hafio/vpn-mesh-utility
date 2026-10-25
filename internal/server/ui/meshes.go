package ui

import (
	"net/http"
	"strings"

	"github.com/hafio/vpn-mesh-utility/internal/server/api"
	"github.com/hafio/vpn-mesh-utility/internal/server/auth"
)

// The Meshes page (amendment A49): the control plane's meshes, to open,
// add, rename and remove. Each change goes through api.Nodes, so it is
// checked against the other meshes and audited as `meshctl mesh` would be.

type meshesData struct {
	Meshes []api.MeshInfo
	// Form is the add form's values, kept on a refusal.
	Form api.MeshAddRequest
}

func (u *ui) meshesPage(w http.ResponseWriter, r *http.Request, s auth.Session) {
	u.renderMeshes(w, r, s, http.StatusOK, page{}, api.MeshAddRequest{})
}

func (u *ui) renderMeshes(w http.ResponseWriter, r *http.Request, s auth.Session, status int, p page, form api.MeshAddRequest) {
	l, err := u.Nodes.Meshes(r.Context())
	if err != nil {
		u.fail(w, r, s, err)
		return
	}
	p.Data = meshesData{Meshes: l.Meshes, Form: form}
	u.show(w, r, s, status, "meshes", "meshes", "Meshes", p)
}

func (u *ui) addMesh(w http.ResponseWriter, r *http.Request, s auth.Session) {
	form := api.MeshAddRequest{Name: strings.TrimSpace(r.PostFormValue("name")), MeshCIDR: strings.TrimSpace(r.PostFormValue("mesh_cidr")),
		Zone: strings.TrimSpace(r.PostFormValue("dns_zone")), IPv6: strings.TrimSpace(r.PostFormValue("ipv6"))}
	up, err := u.Nodes.AddMesh(r.Context(), actor(s), form)
	if err != nil {
		status, msg := message(err)
		u.renderMeshes(w, r, s, status, page{Error: msg}, form)
		return
	}
	m := up.Mesh
	u.renderMeshes(w, r, s, http.StatusOK, page{Flash: "Added mesh " + m.Name + " on " + m.MeshCIDR + ", zone " + m.Zone +
		"; open it to add its nodes"}, api.MeshAddRequest{})
}

func (u *ui) renameMesh(w http.ResponseWriter, r *http.Request, s auth.Session) {
	name := strings.TrimSpace(r.PostFormValue("name"))
	up, err := u.Nodes.SetMesh(r.Context(), actor(s), r.PathValue("name"), api.MeshSetRequest{Name: &name})
	if err != nil {
		status, msg := message(err)
		u.renderMeshes(w, r, s, status, page{Error: msg}, api.MeshAddRequest{})
		return
	}
	u.renderMeshes(w, r, s, http.StatusOK, page{Flash: "Renamed the mesh to " + up.Mesh.Name + "; its uuid and slot stay"}, api.MeshAddRequest{})
}

func (u *ui) removeMesh(w http.ResponseWriter, r *http.Request, s auth.Session) {
	name := r.PathValue("name")
	if err := u.Nodes.RemoveMesh(r.Context(), actor(s), name); err != nil {
		status, msg := message(err)
		u.renderMeshes(w, r, s, status, page{Error: msg}, api.MeshAddRequest{})
		return
	}
	u.renderMeshes(w, r, s, http.StatusOK, page{Flash: "Removed mesh " + name + "; its nightly exports stay on disk"}, api.MeshAddRequest{})
}
