(defpurefun ((vanishes! :@loob) x) x)
(defcolumns (X :i16@prove))
(defpermutation (Y) ((↓ X)))
(defpermutation (Z) ((+ X)))
;; Y == Z
(defconstraint eq () (vanishes! (- Y Z)))
