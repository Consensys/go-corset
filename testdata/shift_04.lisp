(defpurefun ((vanishes! :@loob) x) x)

(defcolumns X Y ST)
(defconstraint c1 () (vanishes! (* ST (- (shift X 1) Y))))
