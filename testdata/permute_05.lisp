(defcolumns
  (X :byte@loob@prove)
  (Y :byte@loob@prove))
(defpermutation (A B) ((+ X) (+ Y)))
(defconstraint diag_ab () (- (shift A 1) B))
