(defpurefun (vanishes! x) (== 0 x))
(defcolumns (X :i16@prove))
(defpermutation (Y) ((↓ X)))
(defpermutation (Z) ((+ X)))
;; Y == Z
(defconstraint eq () (vanishes! (- Y Z)))
