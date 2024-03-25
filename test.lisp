(module test)
(defcolumns A B)
(defpurefun ((and2 :i8) (a b)) (- (+ a b) (* a b)))
