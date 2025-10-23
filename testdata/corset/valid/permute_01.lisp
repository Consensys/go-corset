(defcolumns (X :i16@prove))
(defpermutation (Y) ((↓ X)))
(defpermutation (Z) ((+ X)))
;; Y == Z
(defconstraint eq () (== 0 (- Y Z)))
