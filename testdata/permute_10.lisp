(defcolumns
  (X :byte@prove)
  (Y :byte@prove))
(defpermutation (A B) ((+ X) Y))
(defconstraint diag_ab () (== (shift A 1) B))
