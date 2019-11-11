package cache

func (c *Cache) PurgeClientCache(clientID string) (int32, error) {

	deleteIDs, err := c.CreateID(":authc:*:", clientID, ":*")
	if err != nil {
		return 0, err
	}
	keysResult := c.Client.Keys(deleteIDs.String())

	err = keysResult.Err()
	if err != nil {
		return 0, err
	}

	keys := keysResult.Val()

	if len(keys) > 0 {
		delResult := c.Client.Del(keys...)

		err = delResult.Err()
		if err != nil {
			return 0, err
		}

		return int32(delResult.Val()), nil
	}

	deleteIDs, err = c.CreateID(":tkn:", clientID, ":*")
	if err != nil {
		return 0, err
	}
	keysResult = c.Client.Keys(deleteIDs.String())

	err = keysResult.Err()
	if err != nil {
		return 0, err
	}

	keys = keysResult.Val()

	if len(keys) > 0 {
		delResult := c.Client.Del(keys...)

		err = delResult.Err()
		if err != nil {
			return 0, err
		}

		return int32(delResult.Val()), nil
	}

	return 0, nil

}

func (c *Cache) PurgeSubjectCache(subject string, sessionID string) (int32, error) {

	if len(sessionID) == 0 {
		sessionID = "*"
	}

	var count int32 = 0

	deleteIDs, err := c.CreateID(":sess:", sessionID, ":", subject, '*')
	if err != nil {
		return count, err
	}
	keysResult := c.Client.Keys(deleteIDs.String())

	err = keysResult.Err()
	if err != nil {
		return count, err
	}

	keys := keysResult.Val()

	if len(keys) > 0 {
		delResult := c.Client.Del(keys...)

		err = delResult.Err()
		if err != nil {
			return count, err
		}

		count += int32(delResult.Val())
	}

	deleteIDs, err = c.CreateID(":tkn:*:", subject, ":*")
	if err != nil {
		return count, err
	}
	keysResult = c.Client.Keys(deleteIDs.String())

	err = keysResult.Err()
	if err != nil {
		return count, err
	}

	keys = keysResult.Val()

	if len(keys) > 0 {
		delResult := c.Client.Del(keys...)

		err = delResult.Err()
		if err != nil {
			return count, err
		}

		count += int32(delResult.Val())
	}

	if sessionID != "*" {
		deleteIDs, err = c.CreateID(":act:", sessionID, ":*")
		if err != nil {
			return count, err
		}
		keysResult := c.Client.Keys(deleteIDs.String())

		err = keysResult.Err()
		if err != nil {
			return count, err
		}

		keys = keysResult.Val()

		if len(keys) > 0 {
			delResult := c.Client.Del(keys...)

			err = delResult.Err()
			if err != nil {
				return count, err
			}

			count += int32(delResult.Val())
		}
	}

	return count, nil

}
