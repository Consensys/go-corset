(defcolumns (X :u16))
(defpermutation (Y) ((↓ X)))
(defpermutation (Z) ((+ X)))
;; Y == Z
(defconstraint eq () (- Y Z))
