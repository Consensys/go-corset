(defpurefun ((vanishes! :@loob) x) x)
(defcolumns (X :byte@prove) (Y :byte@prove))
;;
(module m1)
(defcolumns (X :byte@prove) (Y :byte@prove))
(defpermutation (A B) ((+ X) (+ Y)))
(defconstraint diag_ab () (vanishes! (- (shift A 1) B)))
