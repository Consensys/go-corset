(defpurefun ((vanishes! :𝔽@loob) x) x)
;;
(defcolumns (P :binary@prove))
(defperspective p1 P ((X :i16@prove)))
(defpermutation (Y) ((↓ p1/X)))
(defpermutation (Z) ((+ p1/X)))
;; Y == Z
(defconstraint eq () (vanishes! (- Y Z)))
