(defcolumns
  (X :u16)
  (Y :u16))
(defpermutation (A B) ((+ X) (+ Y)))
(defconstraint diag_ab () (- (shift A 1) B))
