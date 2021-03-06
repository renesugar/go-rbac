package gorbac

import "sync"

var empty = struct{}{}

type _RBAC struct {
	mutex       sync.RWMutex
	roles       Roles
	permissions Permissions
	parents     map[string]map[string]struct{}
}

// SetParents bind `parents` to the role `id`.
// If the role or any of parents is not existing,
// an error will be returned.
func (rbac *_RBAC) SetParents(id string, parents []string) error {
	rbac.mutex.Lock()
	defer rbac.mutex.Unlock()
	if _, ok := rbac.roles[id]; !ok {
		return ErrRoleNotExist
	}
	for _, parent := range parents {
		if _, ok := rbac.roles[parent]; !ok {
			return ErrRoleNotExist
		}
	}
	if _, ok := rbac.parents[id]; !ok {
		rbac.parents[id] = make(map[string]struct{})
	}
	for _, parent := range parents {
		rbac.parents[id][parent] = empty
	}
	return nil
}

// GetParents return `parents` of the role `id`.
// If the role is not existing, an error will be returned.
// Or the role doesn't have any parents,
// a nil slice will be returned.
func (rbac *_RBAC) GetParents(id string) ([]string, error) {
	rbac.mutex.Lock()
	defer rbac.mutex.Unlock()
	if _, ok := rbac.roles[id]; !ok {
		return nil, ErrRoleNotExist
	}
	ids, ok := rbac.parents[id]
	if !ok {
		return nil, nil
	}
	var parents []string
	for parent := range ids {
		parents = append(parents, parent)
	}
	return parents, nil
}

// SetParent bind the `parent` to the role `id`.
// If the role or the parent is not existing,
// an error will be returned.
func (rbac *_RBAC) SetParent(id string, parent string) error {
	rbac.mutex.Lock()
	defer rbac.mutex.Unlock()
	if _, ok := rbac.roles[id]; !ok {
		return ErrRoleNotExist
	}
	if _, ok := rbac.roles[parent]; !ok {
		return ErrRoleNotExist
	}
	if _, ok := rbac.parents[id]; !ok {
		rbac.parents[id] = make(map[string]struct{})
	}
	var empty struct{}
	rbac.parents[id][parent] = empty
	return nil
}

// RemoveParent unbind the `parent` with the role `id`.
// If the role or the parent is not existing,
// an error will be returned.
func (rbac *_RBAC) RemoveParent(id string, parent string) error {
	rbac.mutex.Lock()
	defer rbac.mutex.Unlock()
	if _, ok := rbac.roles[id]; !ok {
		return ErrRoleNotExist
	}
	if _, ok := rbac.roles[parent]; !ok {
		return ErrRoleNotExist
	}
	delete(rbac.parents[id], parent)
	return nil
}

// Add a role `r`.
func (rbac *_RBAC) AddRole(r Role) error {
	rbac.mutex.Lock()
	defer rbac.mutex.Unlock()
	if r != nil {
		rbac.roles[r.ID()] = r
	}
	return nil
}

// Remove the role by `id`.
func (rbac *_RBAC) RemoveRole(id string) error {
	rbac.mutex.Lock()
	defer rbac.mutex.Unlock()
	if _, ok := rbac.roles[id]; !ok {
		return ErrRoleNotExist
	}
	delete(rbac.roles, id)
	for rid, parents := range rbac.parents {
		if rid == id {
			delete(rbac.parents, rid)
			continue
		}
		for parent := range parents {
			if parent == id {
				delete(rbac.parents[rid], id)
				break
			}
		}
	}
	return nil
}

// Get the role by `id` and a slice of its parents id.
func (rbac *_RBAC) GetRole(id string) (Role, []string, error) {
	rbac.mutex.RLock()
	defer rbac.mutex.RUnlock()
	r, ok := rbac.roles[id]
	if !ok {
		return nil, nil, ErrRoleNotExist
	}
	var parents []string
	for parent := range rbac.parents[id] {
		parents = append(parents, parent)
	}
	return r, parents, nil
}

// Get the role by `id`.
func (rbac *_RBAC) GetRoleOnly(id string) (Role, error) {
	rbac.mutex.RLock()
	defer rbac.mutex.RUnlock()
	r, ok := rbac.roles[id]
	if !ok {
		return nil, ErrRoleNotExist
	}
	return r, nil
}

// IsGranted tests if the role `id` has Permission `p`.
func (rbac *_RBAC) IsGranted(id string, p Permission) bool {
	return rbac.IsAssertGranted(id, p, nil)
}

// IsAssertGranted tests if the role `id` has Permission `p` with the condition `assert`.
func (rbac *_RBAC) IsAssertGranted(id string, p Permission, assert AssertionFunc) bool {
	rbac.mutex.RLock()
	defer rbac.mutex.RUnlock()
	return rbac.isGranted(id, p, assert)
}

func (rbac *_RBAC) isGranted(id string, p Permission, assert AssertionFunc) bool {
	if assert != nil && !assert(rbac, id, p) {
		return false
	}
	return rbac.recursionCheckID(id, p.ID())
}

// IsGranted tests if the role `id` has Permission id `pid`.
func (rbac *_RBAC) IsGrantedID(id, pid string) bool {
	return rbac.IsAssertGrantedID(id, pid, nil)
}

// IsAssertGranted tests if the role `id` has Permission id `pid` with the condition `assert`.
func (rbac *_RBAC) IsAssertGrantedID(id, pid string, assert AssertionIDFunc) bool {
	rbac.mutex.RLock()
	defer rbac.mutex.RUnlock()
	return rbac.isGrantedID(id, pid, assert)
}

func (rbac *_RBAC) isGrantedID(id, pid string, assert AssertionIDFunc) bool {
	if assert != nil && !assert(rbac, id, pid) {
		return false
	}
	return rbac.recursionCheckID(id, pid)
}

func (rbac *_RBAC) recursionCheckID(id, pid string) bool {
	if role, ok := rbac.roles[id]; ok {
		if role.PermitID(pid) {
			return true
		}
		if parents, ok := rbac.parents[id]; ok {
			for pID := range parents {
				if _, ok := rbac.roles[pID]; ok {
					if rbac.recursionCheckID(pID, pid) {
						return true
					}
				}
			}
		}
	}
	return false
}
